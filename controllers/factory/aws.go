package factory

import (
	"context"
	"github.com/agill17/s3-operator/api/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"math"
	"os"
)

const (
	EnvVarS3Endpoint = "S3_ENDPOINT" // used for mock s3 server to do integration testing
)

type awsClient struct {
	s3Client s3iface.S3API
}

func NewS3(region string) *awsClient {
	cfg := &aws.Config{
		Region:                        aws.String(region),
		CredentialsChainVerboseErrors: aws.Bool(true),
		MaxRetries:                    aws.Int(math.MaxInt64),
	}
	if val, ok := os.LookupEnv(EnvVarS3Endpoint); ok {
		cfg.Endpoint = aws.String(val)
		cfg.DisableSSL = aws.Bool(true)
		cfg.S3ForcePathStyle = aws.Bool(true)
	}

	sess := session.Must(session.NewSession())
	client := s3.New(sess, cfg)
	return &awsClient{s3Client: client}
}

func (a *awsClient) BucketExists(name string) (bool, error) {
	_, err := a.s3Client.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: aws.String(name)})
	if err != nil {
		if awsErr, isAwsErr := err.(awserr.Error); isAwsErr && awsErr.Code() == s3.ErrCodeNoSuchBucket {
			return false, nil
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

func (a *awsClient) ApplyBucketProperties(cr *v1alpha1.Bucket) error {

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
