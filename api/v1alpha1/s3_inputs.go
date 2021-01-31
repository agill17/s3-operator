package v1alpha1

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (b *Bucket) CreateBucketIn() *s3.CreateBucketInput {
	in := &s3.CreateBucketInput{
		Bucket:                     aws.String(b.Spec.BucketName),
		ObjectLockEnabledForBucket: aws.Bool(b.Spec.EnableObjectLock),
	}
	if b.Spec.Region != "us-east-1" {
		in.CreateBucketConfiguration = &s3.CreateBucketConfiguration{LocationConstraint: aws.String(b.Spec.Region)}
	}

	return in
}

func (b *Bucket) DeleteBucketIn() *s3.DeleteBucketInput {
	return &s3.DeleteBucketInput{
		Bucket: aws.String(b.Spec.BucketName),
	}
}

func (b *Bucket) PutBucketVersioningIn() *s3.PutBucketVersioningInput {
	var status = s3.BucketVersioningStatusSuspended
	if b.Spec.EnableVersioning {
		status = s3.BucketVersioningStatusEnabled
	}
	return &s3.PutBucketVersioningInput{
		Bucket: aws.String(b.Spec.BucketName),
		VersioningConfiguration: &s3.VersioningConfiguration{
			Status: aws.String(status),
		},
	}
}

func (b *Bucket) BucketAccelerationConfigIn() *s3.PutBucketAccelerateConfigurationInput {
	var status = s3.BucketAccelerateStatusSuspended
	if b.Spec.EnableTransferAcceleration {
		status = s3.BucketAccelerateStatusEnabled
	}
	in := &s3.PutBucketAccelerateConfigurationInput{
		AccelerateConfiguration: &s3.AccelerateConfiguration{Status: aws.String(status)},
		Bucket:                  aws.String(b.Spec.BucketName),
	}
	return in
}

func (b *Bucket) BucketPolicyInput() *s3.PutBucketPolicyInput {
	in := &s3.PutBucketPolicyInput{
		Bucket: aws.String(b.Spec.BucketName),
		Policy: aws.String(b.Spec.BucketPolicy),
	}
	return in
}

func (b *Bucket) PutTagsIn(tags *s3.Tagging) *s3.PutBucketTaggingInput {

	return &s3.PutBucketTaggingInput{
		Bucket:  aws.String(b.Spec.BucketName),
		Tagging: tags,
	}
}
