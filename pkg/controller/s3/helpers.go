package s3

import (
	"context"
	"github.com/agill17/s3-operator/pkg/apis/agill/v1alpha1"
	"github.com/agill17/s3-operator/pkg/utils"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func secretAccessKeyAndIamAccessKeyMatch(cr *v1alpha1.S3, k8sSecret *v1.Secret, iamClient iamiface.IAMAPI) (bool, error) {
	accessKeyIdInSecret := string(k8sSecret.Data["AWS_ACCESS_KEY_ID"])
	accessKeyIdInAWS, err := utils.GetAccessKeyForUser(cr.Spec.IAMUserSpec.Username, iamClient)
	if err != nil {
		return false, err
	}

	return accessKeyIdInSecret == accessKeyIdInAWS, nil
}

func getIamK8sSecret(cr *v1alpha1.S3, client2 client.Client) (*v1.Secret, error) {
	s := &v1.Secret{}
	err := client2.Get(context.TODO(), types.NamespacedName{Name: cr.GetIAMK8SSecretName(), Namespace: cr.GetNamespace()}, s)
	return s, err
}