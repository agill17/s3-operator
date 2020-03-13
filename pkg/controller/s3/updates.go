package s3

import (
	"context"
	"github.com/agill17/s3-operator/pkg/apis/agill/v1alpha1"
	"github.com/agill17/s3-operator/pkg/controller/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// TODO: need a better way to check if reconcile is needed, this is not extendable
func (r ReconcileS3) isReconcileNeeded(cr *v1alpha1.S3) (bool, error) {
	userExists, err := utils.IAMUserExists(cr.Spec.IAMUserSpec.Username, r.iamClient)
	if err != nil {
		return false, err
	}
	if !userExists {
		return true, nil
	}

	// if secret no longer exists, we reconcile
	k8sSecret, errGettingSecret := getIamK8sSecret(cr, r.client)
	if errGettingSecret != nil {
		if errors.IsNotFound(errGettingSecret) {
			return true, nil
		}
		return false, err
	}

	// if secret does exist and access key no longer matches, we requeue
	accessKeyMatchesInAws, errCheckingMatch := secretAccessKeyAndIamAccessKeyMatch(cr,
		k8sSecret, r.iamClient)
	if errCheckingMatch != nil {
		return false, errCheckingMatch
	}
	if !accessKeyMatchesInAws {
		return true, nil
	}

	// if s3 service no longer exists, we reconcile
	svc := &v1.Service{}
	errGettingSvc := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.GetName(), Namespace:cr.GetNamespace()}, svc)
	if errGettingSvc != nil {
		if errors.IsNotFound(errGettingSvc) {
			return true, nil
		}
		return false, errGettingSvc
	}

	// if s3 bucket no longer exists, we reconcile
	bucketExists,errGettingS3 := utils.BucketExists(cr.Spec.BucketName, r.s3Client)
	if errGettingS3 != nil {
		return false, errGettingS3
	}
	return !bucketExists, nil

}

