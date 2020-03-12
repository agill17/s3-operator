package s3

import (
	agillv1alpha1 "github.com/agill17/s3-operator/pkg/apis/agill/v1alpha1"
	"github.com/agill17/s3-operator/pkg/controller/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func handleEmptyPhase (cr *agillv1alpha1.S3, client client.Client) (reconcile.Result, error) {
	if cr.Status.Phase != agillv1alpha1.CREATE_CLOUD_RESOURCES {
		cr.Status.Phase = agillv1alpha1.CREATE_CLOUD_RESOURCES
		if err := utils.UpdateCrStatus(cr, client); err != nil {
			return reconcile.Result{}, err
		}
	}
	// always requeue in this phase to go to next phase
	return reconcile.Result{Requeue:true}, nil
}

// meant to create cloud resources if they do not exist ( s3, iam user )
func (r ReconcileS3) handleCreateCloudResources(cr *agillv1alpha1.S3) (reconcile.Result, error) {

	// create iam user
	if errCreatingIamResources := r.createIamResources(cr); errCreatingIamResources!= nil {
		return reconcile.Result{}, errCreatingIamResources
	}

	// create bucket
	if errCreatingBucket := r.createBucket(cr); errCreatingBucket != nil {
		return reconcile.Result{}, errCreatingBucket
	}

	// change phase to create k8s resources
	cr.Status.Phase = agillv1alpha1.CREATE_K8S_RESOURCES
	return reconcile.Result{Requeue:true}, utils.UpdateCrStatus(cr, r.client)
}

func (r ReconcileS3) handleCreateK8sResources(cr *agillv1alpha1.S3) (reconcile.Result, error) {

	if errCreatingSecret := createIamK8sSecret(cr, r.client, r.scheme); errCreatingSecret != nil {
		return reconcile.Result{}, errCreatingSecret
	}

	if errCreatingSvc := createS3K8sService(cr, r.client, r.scheme); errCreatingSvc != nil {
		return reconcile.Result{}, errCreatingSvc
	}

	cr.Status.Phase = agillv1alpha1.COMPLETED
	return reconcile.Result{Requeue:true}, utils.UpdateCrStatus(cr, r.client)
}
