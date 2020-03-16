package s3

import (
	agillv1alpha1 "github.com/agill17/s3-operator/pkg/apis/agill/v1alpha1"
	"github.com/agill17/s3-operator/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

func (r ReconcileS3) handleCreateIamResources(cr *agillv1alpha1.S3) error {
	// create iam user
	errCreatingIamUser := utils.CreateIAMUser(cr.CreateIAMUserIn(), r.iamClient)
	if errCreatingIamUser != nil {
		return errCreatingIamUser
	}

	if errCreatingUpdatingPolicy := CreateOrUpdateIAMPolicy(cr, r.iamClient); errCreatingUpdatingPolicy != nil {
		return errCreatingUpdatingPolicy
	}

	return handleAccessKeys(cr, r.iamClient, r.client, r.scheme);
}

// meant to create cloud resources if they do not exist ( s3, iam user )
func (r ReconcileS3) handleCreateS3Resources(cr *agillv1alpha1.S3) error {

	// create bucket
	if errCreatingBucket := r.createBucket(cr); errCreatingBucket != nil {
		return errCreatingBucket
	}

	// change phase to completed
	r.recorder.Eventf(cr, v1.EventTypeNormal, "COMPLETED", "All resources are successfully reconciled.")
	return createS3K8sService(cr, r.client, r.scheme)
}


