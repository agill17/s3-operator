package factory

import (
	"github.com/agill17/s3-operator/api/v1alpha1"
	"github.com/agill17/s3-operator/internal/factory/aws"
	"github.com/agill17/s3-operator/internal/factory/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/**
TODO: Create a a generic type for taking inputs needed for create,delete
Then create a map func per implementation that uses generic type and maps to implementation type for inputs
*/
type Bucket interface {
	BucketExists(in *v1alpha1.Bucket) (bool, error)
	CreateBucket(in *v1alpha1.Bucket) error
	DeleteBucket(in *v1alpha1.Bucket) error
	ApplyBucketProperties(in *v1alpha1.Bucket) error
}

// TODO: extend to other providers
func NewBucket(c client.Client, provider *v1alpha1.Provider) (Bucket, error) {
	switch provider.Name {
	case v1alpha1.Aws:
		return aws.NewS3(c, provider.Region, provider.SecretRef)
	case v1alpha1.Gcp:
		return gcp.NewGCPBucket(c, provider.SecretRef)
	}
	return aws.NewS3(c, provider.Region, provider.SecretRef)
}
