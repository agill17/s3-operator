package s3

import (
	"context"
	"github.com/agill17/s3-operator/pkg/controller/utils"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/davecgh/go-spew/spew"
	"k8s.io/client-go/tools/record"

	agillv1alpha1 "github.com/agill17/s3-operator/pkg/apis/agill/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const S3_CONTROLLER = "s3Controller"
var log = logf.Log.WithName("controller_s3")

// Add creates a new S3 Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileS3{client: mgr.GetClient(), scheme: mgr.GetScheme(), recorder: mgr.GetEventRecorderFor(S3_CONTROLLER)}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("s3-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource S3
	err = c.Watch(&source.Kind{Type: &agillv1alpha1.S3{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileS3 implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileS3{}

// ReconcileS3 reconciles a S3 object
type ReconcileS3 struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	s3Client s3iface.S3API
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a S3 object and makes changes based on the state read
// and what is in the S3.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileS3) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling S3")

	// Fetch the S3 instance
	cr := &agillv1alpha1.S3{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cr)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// add finalizer
	if errAddingFinalizer := utils.AddFinalizer(utils.S3_FINALIZER, r.client, cr); errAddingFinalizer != nil {
		reqLogger.Error(errAddingFinalizer, "Failed to add s3 finalizer, requeue with exponential back-off")
		return reconcile.Result{}, errAddingFinalizer
	}

	// set up s3 client
	if r.s3Client == nil {
		r.s3Client = utils.S3Client(cr.Spec.Region)
	}


	// handle delete
	if cr.GetDeletionTimestamp() != nil {
		// TODO: Add delete logic for s3
		if errRemovingFinalizers := utils.RemoveFinalizer(utils.S3_FINALIZER, cr, r.client); errRemovingFinalizers != nil {
			reqLogger.Error(errRemovingFinalizers, "Failed to remove s3 finalizer, retrying..")
			return reconcile.Result{}, errRemovingFinalizers
		}
	}


	// s3 create and reconcile logic
	if errCreatingBucket := r.createBucket(cr); errCreatingBucket != nil {
		reqLogger.Error(errCreatingBucket, "Failed to create bucket, re-trying..")
		return reconcile.Result{}, errCreatingBucket
	}
	spew.Dump(utils.GetBucketACL(cr.Spec.BucketName, r.s3Client))

	return reconcile.Result{}, nil
}