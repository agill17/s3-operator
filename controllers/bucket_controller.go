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
	"github.com/agill17/s3-operator/pkg/factory"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

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

// +kubebuilder:rbac:groups=agill.apps.agill.apps.s3-operator,resources=buckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agill.apps.agill.apps.s3-operator,resources=buckets/status,verbs=get;update;patch

func (r *BucketReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("bucket", req.NamespacedName)

	namespacedName := fmt.Sprintf("%s/%s", req.Namespace, req.Name)
	cr := &agillappsv1alpha1.Bucket{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, cr)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, fmt.Sprintf("%s: Failed to get bucket CR", namespacedName))
		return ctrl.Result{}, err
	}

	// add finalizer
	if err := FinalizerOp(cr, r.Client, Add, Finalizer); err != nil {
		r.Log.Error(err, fmt.Sprintf("%s: Failed to add finalizer to CR", err))
		return ctrl.Result{}, err
	}

	bucketInterface := factory.NewBucket(cr.Spec.Region)

	// Handle Delete
	if cr.GetDeletionTimestamp() != nil {
		if errDeleting := bucketInterface.DeleteBucket(cr.DeleteBucketIn()); errDeleting != nil {
			r.Log.Error(errDeleting, fmt.Sprintf("%s: Failed to delete bucket", errDeleting))
			return ctrl.Result{}, errDeleting
		}

		if errRemovingFinalizer := FinalizerOp(cr, r.Client, Remove, Finalizer); errRemovingFinalizer != nil {
			r.Log.Error(errRemovingFinalizer, fmt.Sprintf("%s: Failed to remove finalizer", errRemovingFinalizer))
		}
		return ctrl.Result{}, nil
	}

	bucketExists, errCheckingBucket := bucketInterface.BucketExists(cr.Spec.BucketName)
	if errCheckingBucket != nil {
		r.Log.Error(errCheckingBucket, fmt.Sprintf("%s: failed to check if bucket exists", errCheckingBucket))
		return ctrl.Result{}, errCheckingBucket
	}

	if !bucketExists {
		if errCreatingBucket := bucketInterface.CreateBucket(cr.CreateBucketIn()); errCreatingBucket != nil {
			r.Log.Error(errCreatingBucket, fmt.Sprintf("%s: Failed to create s3 bucket", errCreatingBucket))
			return ctrl.Result{}, errCreatingBucket
		}
	}

	return ctrl.Result{}, nil
}

func (r *BucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agillappsv1alpha1.Bucket{}).
		Complete(r)
}
