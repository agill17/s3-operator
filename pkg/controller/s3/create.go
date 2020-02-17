package s3

import (
	"github.com/agill17/s3-operator/pkg/apis/agill/v1alpha1"
	"github.com/agill17/s3-operator/pkg/controller/utils"
	"github.com/davecgh/go-spew/spew"
	v1 "k8s.io/api/core/v1"
)

func (r ReconcileS3) createBucket(cr *v1alpha1.S3) error {
	exists, errGettingBucket := utils.BucketExists(cr.Spec.BucketName, r.s3Client)
	if errGettingBucket != nil {
		r.recorder.Eventf(cr, v1.EventTypeWarning, "FAILED", "Failed to get bucket from Cloud: %v", errGettingBucket)
		return errGettingBucket
	}

	if !exists {
		r.recorder.Eventf(cr, v1.EventTypeNormal, "CREATING", "Bucket does not exist, creating now...")
		out, err := r.s3Client.CreateBucket(cr.CreateBucketIn())
		if err != nil {
			r.recorder.Eventf(cr, v1.EventTypeWarning, "FAILED", "Failed to create bucket: %v", err)
			return err
		}
		spew.Dump(out)
		r.recorder.Eventf(cr, v1.EventTypeNormal, "CREATED", "S3 Bucket created successfully")
	}
	return nil
}

