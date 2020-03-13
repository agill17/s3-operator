package s3

import (
	"context"
	"errors"
	"fmt"
	"github.com/agill17/s3-operator/pkg/apis/agill/v1alpha1"
	"github.com/agill17/s3-operator/pkg/controller/utils"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/davecgh/go-spew/spew"
	v1 "k8s.io/api/core/v1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r ReconcileS3) createBucket(cr *v1alpha1.S3) error {
	exists, errGettingBucket := utils.BucketExists(cr.Spec.BucketName, r.s3Client)
	if errGettingBucket != nil {
		r.recorder.Eventf(cr, v1.EventTypeWarning, "FAILED", "Failed to get bucket from Cloud: %v", errGettingBucket)
		return errGettingBucket
	}

	if !exists {
		r.recorder.Eventf(cr, v1.EventTypeNormal, "CREATING", "Bucket does not exist, creating now...")
		out, err := r.s3Client.CreateBucket(cr.CreateBucketIn())
		if err != nil {
			r.recorder.Eventf(cr, v1.EventTypeWarning, "FAILED", "Failed to create bucket: %v", err)
			return err
		}
		spew.Dump(out)
		r.recorder.Eventf(cr, v1.EventTypeNormal, "CREATED", "S3 Bucket created successfully")
	}
	return nil
}

// if secret is not found in namespace, create new access keys ( delete the rest of the access keys if any )
// if secret is found, and access key does not match IAM access key ( delete the secret and delete all access keys on IAM ) and create fresh access keys
func handleAccessKeys(cr *v1alpha1.S3, iamClient iamiface.IAMAPI, client client.Client, scheme *runtime.Scheme) error {
	secretFound := &v1.Secret{}
	if err := client.Get(context.TODO(),
		types.NamespacedName{Namespace: cr.GetNamespace(), Name:fmt.Sprintf("%v-iam-secret", cr.Name)},
		secretFound); err != nil {

			if apierror.IsNotFound(err) {
				// clean up access keys if any
				if errDeletingAllAccessKeys := utils.DeleteAllAccessKeys(cr.Spec.IAMUserSpec.Username, iamClient); errDeletingAllAccessKeys != nil {
					return errDeletingAllAccessKeys
				}

				// create fresh access keys
				acccessKeysOutput, errCreatingAccessKeys := utils.CreateAccessKeys(cr.Spec.IAMUserSpec.Username, iamClient)
				if errCreatingAccessKeys != nil {
					return errCreatingAccessKeys
				}

				// create k8s secret
				if errCreatingSecret := createIamK8sSecret(cr,
					*acccessKeysOutput.AccessKey.AccessKeyId,
					*acccessKeysOutput.AccessKey.SecretAccessKey,
					client, scheme); errCreatingSecret != nil {
					return errCreatingSecret
				}
				return utils.UpdateCrStatus(cr, client)
			}
		// if err is something else other then isNotFound, return that error
		return err
	}

	// if secret is found make sure access keys matches the one in IAM
	accessKeyIdInSecret := string(secretFound.Data["AWS_ACCESS_KEY_ID"])
	accessKeyIdInAWS, err := utils.GetAccessKeyForUser(cr.Spec.IAMUserSpec.Username, iamClient)
	if err != nil {
		return err
	}

	if accessKeyIdInSecret != accessKeyIdInAWS {
		// delete secret to re-initiate
		if err := client.Delete(context.TODO(), secretFound); err != nil {
			return err
		}
		// return error to force a requeue
		// TODO: make this a custom error so that we can catch in reconcile and quietly requeue
		return errors.New("ErrorIAMK8SSecretNeedsUpdate")
	}
	return nil
}

