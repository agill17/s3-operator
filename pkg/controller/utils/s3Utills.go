package utils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"math"
)

func S3Client(region string) s3iface.S3API {
	sess, _ := session.NewSession(&aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
		Region: aws.String(region),
		MaxRetries: aws.Int(math.MaxInt64),
	})
	return s3.New(sess)
}

func BucketExists(bucketName string, s3Client s3iface.S3API) (bool, error) {
	_, err := s3Client.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: &bucketName})
	if awserr, ok := err.(awserr.Error); ok && awserr.Code() == s3.ErrCodeNoSuchBucket {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func GetBucketACL(bucketName string, s3Client s3iface.S3API) (*s3.GetBucketAclOutput, error) {
	bucketAcl, err := s3Client.GetBucketAcl(&s3.GetBucketAclInput{Bucket:aws.String(bucketName)})
	if err != nil {
		return nil, err
	}
	return bucketAcl, nil
}
