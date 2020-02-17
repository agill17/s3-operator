package v1alpha1

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (s S3) CreateBucketIn() *s3.CreateBucketInput {
	s3Input := &s3.CreateBucketInput{
		ACL:                        aws.String(s.Spec.BucketACL),
		Bucket:                     aws.String(s.Spec.BucketName),
		CreateBucketConfiguration:  s.SetBucketLocation(),
		ObjectLockEnabledForBucket: aws.Bool(s.Spec.EnableObjectLock),
	}
	return s3Input
}

func (s S3) DeleteBucketIn() *s3.DeleteBucketInput {
	return &s3.DeleteBucketInput{Bucket:aws.String(s.Spec.BucketName)}
}

func (s S3) SetBucketLocation() *s3.CreateBucketConfiguration {
	if s.Spec.Region != "" {
		return &s3.CreateBucketConfiguration{LocationConstraint: aws.String(s.Spec.Region)}
	}
	return nil
}
