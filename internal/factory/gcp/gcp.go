package gcp

import (
	"context"
	"github.com/agill17/s3-operator/api/v1alpha1"
	"google.golang.org/api/googleapi"
)

func (g gcpClient) BucketExists(in *v1alpha1.Bucket) (bool, error) {
	_, err := g.storageClient.Bucket(in.Spec.BucketName).Attrs(context.TODO())
	if err != nil {
		if gcpErr, ok := err.(*googleapi.Error); ok && gcpErr.Code == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (g gcpClient) CreateBucket(in *v1alpha1.Bucket) error {
	panic("implement me")
}

func (g gcpClient) DeleteBucket(in *v1alpha1.Bucket) error {
	panic("implement me")
}

func (g gcpClient) ApplyBucketProperties(in *v1alpha1.Bucket) error {
	panic("implement me")
}