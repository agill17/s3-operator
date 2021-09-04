package v1alpha1

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (b *Bucket) AWSCreateBucketIn() *s3.CreateBucketInput {
	in := &s3.CreateBucketInput{
		Bucket:                     aws.String(b.Spec.BucketName),
		ObjectLockEnabledForBucket: aws.Bool(b.Spec.EnableObjectLock),
	}
	if b.Spec.Provider.Region != "us-east-1" {
		in.CreateBucketConfiguration = &s3.CreateBucketConfiguration{LocationConstraint: aws.String(b.Spec.Provider.Region)}
	}

	return in
}

func (b *Bucket) AWSDeleteBucketIn() *s3.DeleteBucketInput {
	return &s3.DeleteBucketInput{
		Bucket: aws.String(b.Spec.BucketName),
	}
}

func (b *Bucket) AWSPutBucketVersioningIn() *s3.PutBucketVersioningInput {
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

func (b *Bucket) AWSBucketAccelerationConfigIn() *s3.PutBucketAccelerateConfigurationInput {
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

func (b *Bucket) AWSBucketPolicyInput() *s3.PutBucketPolicyInput {
	in := &s3.PutBucketPolicyInput{
		Bucket: aws.String(b.Spec.BucketName),
		Policy: aws.String(b.Spec.BucketPolicy),
	}
	return in
}

func (b *Bucket) AWSPutTagsIn(tags *s3.Tagging) *s3.PutBucketTaggingInput {

	return &s3.PutBucketTaggingInput{
		Bucket:  aws.String(b.Spec.BucketName),
		Tagging: tags,
	}
}

func (b *Bucket) AWSPutBucketCannedAclInput() *s3.PutBucketAclInput {
	in := &s3.PutBucketAclInput{
		ACL:    aws.String(b.Spec.CannedBucketAcl),
		Bucket: aws.String(b.Spec.BucketName),
	}
	return in
}
