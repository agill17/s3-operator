package aws

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"math"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)
const (
	awsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	awsAccessKeyID = "AWS_ACCESS_KEY_ID"
	EnvVarS3Endpoint = "S3_ENDPOINT" // used for mock s3 server to do integration testing
)

type awsClient struct {
	s3Client s3iface.S3API
}


func NewS3(c client.Client, region string, sRef *v1.SecretReference) (*awsClient, error) {
	creds, err := getAwsCredentials(c, sRef)
	if err != nil {
		return nil, err
	}
	cfg := &aws.Config{
		Region:                        aws.String(region),
		CredentialsChainVerboseErrors: aws.Bool(true),
		MaxRetries:                    aws.Int(math.MaxInt64),
		Credentials: creds,
	}
	if val, ok := os.LookupEnv(EnvVarS3Endpoint); ok {
		cfg.Endpoint = aws.String(val)
		cfg.DisableSSL = aws.Bool(true)
		cfg.S3ForcePathStyle = aws.Bool(true)
	}

	sess := session.Must(session.NewSession())
	client := s3.New(sess, cfg)
	return &awsClient{s3Client: client}, nil
}


/*
	1. Secret Ref takes precedence
	2. Then env vars
*/
func getAwsCredentials(c client.Client, sRef *v1.SecretReference) (*credentials.Credentials, error) {
	if sRef == nil {
		return credentials.NewEnvCredentials(), nil
	}
	secret := &v1.Secret{}
	err := c.Get(context.TODO(), types.NamespacedName{
		Name:      sRef.Name,
		Namespace: sRef.Namespace,
	}, secret)
	if err != nil {
		return nil, err
	}
	return NewK8sSecretCredentialProvider(secret.Data), nil

}
