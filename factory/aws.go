package factory

import (
	"context"
	"fmt"
	"github.com/agill17/s3-operator/api/v1alpha1"
	"github.com/agill17/s3-operator/vault"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"math"
	"os"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"sync"
	"time"
)

const (
	EnvVarMockS3Endpoint = "MOCK_S3_ENDPOINT" // used for mock s3 server to do integration testing
	AwsAccessKeyId       = "AWS_ACCESS_KEY_ID"
	AwsSecretAccessKey   = "AWS_SECRET_ACCESS_KEY"
	VaultRegexp          = `^vault[a-zA-Z0-9\W].*#[a-zA-Z0-9\W].*$`
)

var awsClientCache sync.Map

type awsClient struct {
	s3Client s3iface.S3API
	vClient  *vault.Client
}

func getCacheKeyName(pName, region string) string {
	return fmt.Sprintf("aws-%v-%v", pName, region)
}

//TODO: cache awsClient and delete from cache when "InvalidAccessKeyId" error is thrown.
func NewS3(ctx context.Context, region string, pName string, pCreds map[string][]byte) (*awsClient, error) {
	cacheKeyName := fmt.Sprintf("aws-%v-%v", pName, region)
	cachedAwsClient, ok := awsClientCache.Load(cacheKeyName)
	if !ok {
		cfg := &aws.Config{
			Region:                        aws.String(region),
			CredentialsChainVerboseErrors: aws.Bool(true),
			MaxRetries:                    aws.Int(math.MaxInt64),
		}
		if val, ok := os.LookupEnv(EnvVarMockS3Endpoint); ok {
			cfg.Endpoint = aws.String(val)
			cfg.DisableSSL = aws.Bool(true)
			cfg.S3ForcePathStyle = aws.Bool(true)
		}
		sess := session.Must(session.NewSession())
		client := &awsClient{}
		awsCreds, vaultClient, err := getAwsCreds(pCreds)
		if err != nil {
			return nil, err
		}
		if awsCreds != nil {
			cfg.Credentials = awsCreds
		}
		if vaultClient != nil {
			client.vClient = vaultClient
		}
		client.s3Client = s3.New(sess, cfg)
		awsClientCache.Store(cacheKeyName, client)
		return client, nil
	}

	if time.Now().After(cachedAwsClient.(*awsClient).vClient.ExpectedLeaseToEndAt) {
		awsClientCache.Delete(cacheKeyName)
		return nil, vault.ErrRequeueNeeded{}
	}

	return cachedAwsClient.(*awsClient), nil

}

func getAwsCreds(providerCredsMap map[string][]byte) (*credentials.Credentials, *vault.Client, error) {
	providerAccessId, providerHasAccessID := providerCredsMap[AwsAccessKeyId]
	providerSecretKey, providerHasSecret := providerCredsMap[AwsSecretAccessKey]

	// if a provider does not have the keys we look for in spec.Credentials, let SDK figure out the credentials to use
	if !providerHasAccessID || !providerHasSecret {
		return nil, nil, nil
	}

	// stringify
	strProviderAccessId := string(providerAccessId)
	strProviderSecretKey := string(providerSecretKey)

	vaultRegex, err := regexp.Compile(VaultRegexp)
	if err != nil {
		return nil, nil, err
	}

	// if they look like vault paths, try reading from vault
	if vaultRegex.MatchString(strProviderAccessId) &&
		vaultRegex.MatchString(strProviderSecretKey) {
		vClient, err := vault.NewVaultClient(providerCredsMap)
		if err != nil {
			return nil, nil, err
		}

		// sample vault path
		// vault:kv/data/foo#key
		// remove vault prefix
		strProviderAccessId = strings.Replace(strProviderAccessId, "vault:", "", 1)
		strProviderSecretKey = strings.Replace(strProviderSecretKey, "vault:", "", 1)

		// separate path and key by splitting at #
		accessIDPathKeySplit := strings.Split(strProviderSecretKey, "#")
		secretKeyPathSplit := strings.Split(strProviderSecretKey, "#")

		// TODO: add validation that after split we MUST have EXACTLY 1 element on left and EXACTLY 1 on right == so size == 2

		accessIDFromVault, err := vClient.ReadSecretPath(accessIDPathKeySplit[0], accessIDPathKeySplit[len(accessIDPathKeySplit)-1])
		if err != nil {
			return nil, nil, err
		}
		secretKeyFromVault, err := vClient.ReadSecretPath(secretKeyPathSplit[0], secretKeyPathSplit[len(secretKeyPathSplit)-1])
		if err != nil {
			return nil, nil, err
		}
		return credentials.NewStaticCredentials(accessIDFromVault, secretKeyFromVault, ""), vClient, nil
	}

	// assuming creds are specified in provider.spec.Credentials and they are real aws creds and not vault paths
	return credentials.NewStaticCredentials(string(providerAccessId), string(providerSecretKey), ""), nil, nil
}

func (a *awsClient) BucketExists(name string) (bool, error) {
	_, err := a.s3Client.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: aws.String(name)})
	if err != nil {
		if awsErr, isAwsErr := err.(awserr.Error); isAwsErr {
			if awsErr.Code() == s3.ErrCodeNoSuchBucket {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (a *awsClient) CreateBucket(input *s3.CreateBucketInput) error {
	_, err := a.s3Client.CreateBucket(input)
	return err
}

func (a *awsClient) DeleteBucket(input *s3.DeleteBucketInput) error {
	exists, err := a.BucketExists(*input.Bucket)
	if err != nil {
		return err
	}

	if exists {
		iter := s3manager.NewDeleteListIterator(a.s3Client, &s3.ListObjectsInput{
			Bucket: input.Bucket,
		})

		// Traverse iterator deleting each object
		if err := s3manager.NewBatchDeleteWithClient(a.s3Client).Delete(context.TODO(), iter); err != nil {
			return err
		}

		if _, err := a.s3Client.DeleteBucket(input); err != nil {
			return err
		}
	}
	return nil
}

func (a *awsClient) ApplyBucketProperties(object client.Object) error {
	cr, isBucket := object.(*v1alpha1.Bucket)
	if !isBucket {
		return ErrIsNotBucketObject{Message: fmt.Sprintf("Expected: Bucket but got: %T", object)}
	}

	if _, errApplyingVersioning := a.s3Client.PutBucketVersioning(cr.PutBucketVersioningIn()); errApplyingVersioning != nil {
		return errApplyingVersioning
	}

	if _, errApplyingTACL := a.s3Client.PutBucketAccelerateConfiguration(cr.BucketAccelerationConfigIn()); errApplyingTACL != nil {
		return errApplyingTACL
	}

	policyIn := cr.BucketPolicyInput()
	if errFailedValidation := policyIn.Validate(); errFailedValidation != nil {
		return errFailedValidation
	}
	if _, err := a.s3Client.PutBucketPolicy(policyIn); err != nil {
		return err
	}

	if _, err := a.s3Client.PutBucketAcl(cr.PutBucketCannedAclInput()); err != nil {
		return err
	}

	if _, err := a.s3Client.PutBucketTagging(cr.PutTagsIn(MapToTagging(cr.Spec.Tags))); err != nil {
		return err
	}

	return nil
}

func MapToTagging(m map[string]string) *s3.Tagging {
	t := &s3.Tagging{TagSet: []*s3.Tag{}}
	for k, v := range m {
		t.TagSet = append(t.TagSet, &s3.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	return t
}

func TSetToMap(tSet []*s3.Tag) map[string]string {
	m := map[string]string{}
	for _, ele := range tSet {
		m[*ele.Key] = *ele.Value
	}
	return m

}
