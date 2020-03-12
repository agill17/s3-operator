package s3

import (
	"context"
	"fmt"
	"github.com/agill17/s3-operator/pkg/apis/agill/v1alpha1"
	"github.com/agill17/s3-operator/pkg/controller/utils"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/davecgh/go-spew/spew"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

func (r ReconcileS3) createIamResources(cr *v1alpha1.S3) error {
	errCreatingIamUser := utils.CreateIAMUser(cr.CreateIAMUserIn(), r.iamClient)
	if errCreatingIamUser != nil {
		return errCreatingIamUser
	}

	errAttachingIAMPolicy := utils.AttachPolicyToIAMUser(cr.Spec.IAMUserSpec.Username, cr.Spec.IAMUserSpec.AccessPolicy, r.iamClient)
	if errAttachingIAMPolicy != nil {
		return errAttachingIAMPolicy
	}


	return handleAccessKeys(cr, r.iamClient, r.client)
}

// if secret is not found in namespace, create new access keys ( delete the rest of the access keys if any )
// if secret is found, and access key does not match IAM access key ( delete the secret and delete all access keys on IAM ) and create fresh access keys
func handleAccessKeys(cr *v1alpha1.S3, iamClient iamiface.IAMAPI, client client.Client) error {
	secretFound := &v1.Secret{}
	if err := client.Get(context.TODO(),
		types.NamespacedName{Namespace: cr.GetNamespace(), Name:fmt.Sprintf("%v-iam-secret", cr.Name)},
		secretFound); err != nil {

			if errors.IsNotFound(err) {
				// clean up access keys if any
				if errDeletingAllAccessKeys := utils.DeleteAllAccessKeys(cr.Spec.IAMUserSpec.Username, iamClient); errDeletingAllAccessKeys != nil {
					return errDeletingAllAccessKeys
				}

				// create fresh access keys
				acccessKeysOutput, errCreatingAccessKeys := utils.CreateAccessKeys(cr.Spec.IAMUserSpec.Username, iamClient)
				if errCreatingAccessKeys != nil {
					return errCreatingAccessKeys
				}

				cr.Status.AccessKeyId = *acccessKeysOutput.AccessKey.AccessKeyId
				cr.Status.SecretAccessKey = *acccessKeysOutput.AccessKey.SecretAccessKey
				return utils.UpdateCrStatus(cr, client)
			}
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
		return client.Delete(context.TODO(), secretFound)
	}
	return nil
}

