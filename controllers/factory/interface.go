package factory

import "github.com/aws/aws-sdk-go/service/s3"

/**
TODO: Create a a generic type for taking inputs needed for create,delete
Then create a map func per implementation that uses generic type and maps to implementation type for inputs
*/
type Bucket interface {
	BucketExists(name string) (bool, error)
	CreateBucket(input *s3.CreateBucketInput) error
	DeleteBucket(input *s3.DeleteBucketInput) error
	PutBucketPolicy(input *s3.PutBucketPolicyInput) error
}

// TODO: extend to other providers
func NewBucket(region string) Bucket {
	return NewS3(region)
}
