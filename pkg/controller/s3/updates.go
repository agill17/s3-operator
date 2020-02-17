package s3

import (
	"github.com/agill17/s3-operator/pkg/apis/agill/v1alpha1"
)

func (r *ReconcileS3) reconcileState(cr *v1alpha1.S3) error {
	return nil
}

//func (r *ReconcileS3) updateBucketACLIfNeeded(cr *v1alpha1.S3) error {
//	currentBucketAcl := utils.GetBucketACL(cr.Spec.BucketName, r.s3Client)
//	desiredAcl := cr.Spec.BucketACL
//}