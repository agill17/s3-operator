package factory

import (
	"context"
	"github.com/aws/aws-sdk-go/service/s3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/**
TODO: Create a a generic type for taking inputs needed for create,delete
Then create a map func per implementation that uses generic type and maps to implementation type for inputs
*/
type Bucket interface {
	BucketExists(name string) (bool, error)
	CreateBucket(input *s3.CreateBucketInput) error
	DeleteBucket(input *s3.DeleteBucketInput) error
	ApplyBucketProperties(clientObj client.Object) error
}

// TODO: extend to other providers
func NewBucketInterface(ctx context.Context, providerType, region string, providerCreds map[string][]byte) (Bucket, error) {

	switch providerType {
	case "aws":
		return NewS3(ctx, region, providerCreds)
	}
	return nil, nil
}
