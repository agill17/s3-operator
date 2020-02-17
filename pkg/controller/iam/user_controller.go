package iam

import (
	"context"
	"github.com/agill17/s3-operator/pkg/apis/agill/v1alpha1"
	"github.com/agill17/s3-operator/pkg/controller/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const IAM_USER_CONTROLLER = "iamUserController"

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}
// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIAMUser{client: mgr.GetClient(), scheme: mgr.GetScheme(), recorder: mgr.GetEventRecorderFor(IAM_USER_CONTROLLER)}
}


// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("iam-user-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource S3
	err = c.Watch(&source.Kind{Type: &v1alpha1.S3{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &v1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.S3{},
	})
	if err != nil {
		return err
	}

	return nil
}
var _ reconcile.Reconciler = &ReconcileIAMUser{}

type ReconcileIAMUser struct {
	client client.Client
	scheme *runtime.Scheme
	recorder record.EventRecorder
}
func (r *ReconcileIAMUser) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	//get cr
	cr := &v1alpha1.S3{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cr)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// add finalizer
	if errAddingFinalizer := utils.AddFinalizer(utils.IAM_FINALIZER, r.client, cr); errAddingFinalizer != nil {
		return reconcile.Result{RequeueAfter: 30}, errAddingFinalizer
	}

	// handle delete
	if cr.GetDeletionTimestamp() != nil {
		// TODO: Delete IAM user
		if errRemovingFinalizer := utils.RemoveFinalizer(utils.IAM_FINALIZER, cr, r.client); errRemovingFinalizer != nil {
			return reconcile.Result{}, errRemovingFinalizer
		}
		return reconcile.Result{}, nil // do not requeue
	}




	return reconcile.Result{}, nil
}


