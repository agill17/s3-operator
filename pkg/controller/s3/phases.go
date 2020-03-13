package s3

import (
	agillv1alpha1 "github.com/agill17/s3-operator/pkg/apis/agill/v1alpha1"
	"github.com/agill17/s3-operator/pkg/controller/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func handleEmptyPhase (cr *agillv1alpha1.S3, client client.Client) (reconcile.Result, error) {
	if cr.Status.Phase != agillv1alpha1.CREATE_IAM_RESOURCES {
		cr.Status.Phase = agillv1alpha1.CREATE_IAM_RESOURCES
		if err := utils.UpdateCrStatus(cr, client); err != nil {
			return reconcile.Result{}, err
		}
	}
	// always requeue in this phase to go to next phase
	return reconcile.Result{Requeue:true}, nil
}

func (r ReconcileS3) handleCreateIamResources(cr *agillv1alpha1.S3) (reconcile.Result, error) {
	// create iam user
	errCreatingIamUser := utils.CreateIAMUser(cr.CreateIAMUserIn(), r.iamClient)
	if errCreatingIamUser != nil {
		return reconcile.Result{}, errCreatingIamUser
	}

	// attach policy
	errAttachingIAMPolicy := utils.AttachPolicyToIAMUser(cr.Spec.IAMUserSpec.Username, cr.Spec.IAMUserSpec.AccessPolicy, r.iamClient)
	if errAttachingIAMPolicy != nil {
		return reconcile.Result{},  errAttachingIAMPolicy
	}

	// create access keys and k8s secret
	if errCreatingAccessKeys := handleAccessKeys(cr, r.iamClient, r.client, r.scheme); errCreatingAccessKeys != nil {
		return reconcile.Result{}, errCreatingAccessKeys
	}

	// set to next phase and requeue
	cr.Status.Phase = agillv1alpha1.CREATE_S3_RESOURCES
	return reconcile.Result{Requeue:true}, utils.UpdateCrStatus(cr, r.client)
}

// meant to create cloud resources if they do not exist ( s3, iam user )
func (r ReconcileS3) handleCreateS3Resources(cr *agillv1alpha1.S3) (reconcile.Result, error) {

	// create bucket
	if errCreatingBucket := r.createBucket(cr); errCreatingBucket != nil {
		return reconcile.Result{}, errCreatingBucket
	}

	if errCreatingSvc := createS3K8sService(cr, r.client, r.scheme); errCreatingSvc != nil {
		return reconcile.Result{}, errCreatingSvc
	}

	// change phase to completed
	cr.Status.Phase = agillv1alpha1.COMPLETED
	return reconcile.Result{Requeue:true}, utils.UpdateCrStatus(cr, r.client)
}
