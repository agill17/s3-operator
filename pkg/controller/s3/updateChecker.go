package s3

import (
	"context"
	"github.com/agill17/s3-operator/pkg/apis/agill/v1alpha1"
	"github.com/agill17/s3-operator/pkg/utils"
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
		r.recorder.Eventf(cr, v1.EventTypeWarning, "UPDATING", "IAM User no longer exists on AWS, re-creating..")
		return true, nil
	}

	// if secret no longer exists, we reconcile
	k8sSecret, errGettingSecret := getIamK8sSecret(cr, r.client)
	if errGettingSecret != nil {
		if errors.IsNotFound(errGettingSecret) {
			r.recorder.Eventf(cr, v1.EventTypeWarning, "UPDATING", "IAM Secret no longer exists in namespace")
			r.recorder.Eventf(cr, v1.EventTypeWarning, "UPDATING", "New IAM access keys will be created and deployed")
			return true, nil
		}
		return false, err
	}

	// if restricted policy no longer matches in AWS
	desiredPolicy, errGettingDesiredPolicy := v1alpha1.DesiredRestrictedPolicyDocForBucket(cr.GetPolicyName(), cr.Spec.BucketName)
	if errGettingDesiredPolicy != nil {
		return false, errGettingDesiredPolicy
	}
	policyUpToDate, errCheckingPolicyWithAws := IAMPolicyMatchesDesiredPolicyDocument(desiredPolicy,
		cr.Spec.IAMUserSpec.Username, cr.GetPolicyName(), r.iamClient)
	if errCheckingPolicyWithAws != nil {
		return false, errCheckingPolicyWithAws
	}
	if !policyUpToDate {
		return true, nil
	}

	// if secret does exist and access key no longer matches, we requeue
	accessKeyMatchesInAws, errCheckingMatch := secretAccessKeyAndIamAccessKeyMatch(cr,
		k8sSecret, r.iamClient)
	if errCheckingMatch != nil {
		return false, errCheckingMatch
	}
	if !accessKeyMatchesInAws {
		r.recorder.Eventf(cr, v1.EventTypeWarning, "UPDATING", "Access key ID no longer matches with AWS")
		r.recorder.Eventf(cr, v1.EventTypeWarning, "UPDATING", "New IAM access keys will be created and deployed")
		return true, nil
	}

	// if s3 service no longer exists, we reconcile
	svc := &v1.Service{}
	errGettingSvc := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.GetName(), Namespace:cr.GetNamespace()}, svc)
	if errGettingSvc != nil {
		if errors.IsNotFound(errGettingSvc) {
			r.recorder.Eventf(cr, v1.EventTypeWarning, "UPDATING", "S3 Service no longer exists in namespace")
			r.recorder.Eventf(cr, v1.EventTypeWarning, "UPDATING", "S3 Service will be re-created")
			return true, nil
		}
		return false, errGettingSvc
	}

	// if s3 bucket no longer exists, we reconcile
	bucketExists,errGettingS3 := utils.BucketExists(cr.Spec.BucketName, r.s3Client)
	if errGettingS3 != nil {
		return false, errGettingS3
	}
	if !bucketExists {
		r.recorder.Eventf(cr, v1.EventTypeWarning, "UPDATING", "S3 Bucket no longer exists in AWS")
		r.recorder.Eventf(cr, v1.EventTypeWarning, "UPDATING", "S3 Bucket will be re-created..")
		return true, nil
	}

	return false, nil
}

