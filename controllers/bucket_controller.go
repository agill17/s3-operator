/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/agill17/s3-operator/internal"
	"github.com/agill17/s3-operator/internal/factory"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agillappsv1alpha1 "github.com/agill17/s3-operator/api/v1alpha1"
)

// BucketReconciler reconciles a Bucket object
type BucketReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	Finalizer = "agill.apps.s3-bucket"
)

// +kubebuilder:rbac:groups=agill.apps.s3-operator,resources=buckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agill.apps.s3-operator,resources=buckets/status,verbs=get;update;patch

func (r *BucketReconciler) Reconcile(ctx context.Context,req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("bucket", req.NamespacedName)

	namespacedName := fmt.Sprintf("%s/%s", req.Namespace, req.Name)
	cr := &agillappsv1alpha1.Bucket{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, cr)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, fmt.Sprintf("%s: Failed to get bucket CR", namespacedName))
		return ctrl.Result{}, err
	}
	// add finalizer
	if err := internal.FinalizerOp(cr, r.Client, internal.Add, Finalizer); err != nil {
		r.Log.Error(err, fmt.Sprintf("%s: Failed to add finalizer to CR", err))
		return ctrl.Result{}, err
	}

	bInterface, err := factory.NewBucket(r.Client, &cr.Spec.Provider)
	if err != nil {
		return ctrl.Result{}, err
	}
	
	// Handle Delete
	if cr.GetDeletionTimestamp() != nil {
		r.Log.Info(fmt.Sprintf("%v: deleting bucket", namespacedName))
		if errDeleting := bInterface.DeleteBucket(cr); errDeleting != nil {
			r.Log.Error(errDeleting, fmt.Sprintf("%s: Failed to delete bucket", errDeleting))
			return ctrl.Result{}, errDeleting
		}

		if errRemovingFinalizer := internal.FinalizerOp(cr, r.Client, internal.Remove, Finalizer); errRemovingFinalizer != nil {
			r.Log.Error(errRemovingFinalizer, fmt.Sprintf("%s: Failed to remove finalizer", errRemovingFinalizer))
		}
		return ctrl.Result{}, nil
	}

	bucketExists, errCheckingBucket := bInterface.BucketExists(cr)
	if errCheckingBucket != nil {
		r.Log.Error(errCheckingBucket, fmt.Sprintf("%s: failed to check if bucket exists", errCheckingBucket))
		return ctrl.Result{}, errCheckingBucket
	}

	if !bucketExists {
		r.Log.Info(fmt.Sprintf("%v: creating bucket", namespacedName))
		if errCreatingBucket := bInterface.CreateBucket(cr); errCreatingBucket != nil {
			r.Log.Error(errCreatingBucket, fmt.Sprintf("%s: Failed to create s3 bucket", errCreatingBucket))
			return ctrl.Result{}, errCreatingBucket
		}
	}

	errApplyingBucketProperties := bInterface.ApplyBucketProperties(cr)
	if errApplyingBucketProperties != nil {
		r.Log.Error(errApplyingBucketProperties, fmt.Sprintf("Failed to apply bucket properties"))
		return ctrl.Result{}, errApplyingBucketProperties
	}

	if !cr.Status.Ready {
		cr.Status.Ready = true
		if err := r.Client.Status().Update(context.TODO(), cr); err != nil {
			r.Log.Error(err, "Failed to update bucket status")
			return ctrl.Result{}, err
		}
	}

	r.Log.Info(fmt.Sprintf("%v: Bucket reconciled", namespacedName))

	return ctrl.Result{}, nil
}

func (r *BucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agillappsv1alpha1.Bucket{}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				return e.ObjectNew.GetGeneration() != e.ObjectOld.GetGeneration()
			},
		}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 10}).
		Complete(r)
}
