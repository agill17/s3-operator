package s3

import (
	"context"
	agillv1alpha1 "github.com/agill17/s3-operator/pkg/apis/agill/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func createIamK8sSecret(cr *agillv1alpha1.S3, accessKeyId, secretAccessKey string, client client.Client, scheme *runtime.Scheme) error {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.GetIAMK8SSecretName(),
			Namespace: cr.GetNamespace(),
		},
		Data: map[string][]byte{
			"AWS_ACCESS_KEY_ID":     []byte(accessKeyId),
			"AWS_SECRET_ACCESS_KEY": []byte(secretAccessKey),
		},
		Type: v1.SecretTypeOpaque,
	}

	if _, err := controllerutil.CreateOrUpdate(context.TODO(), client, secret, func() error {
		return controllerutil.SetControllerReference(cr, secret, scheme)
	}); err != nil {
		return err
	}

	return nil
}

func createS3K8sService(cr *agillv1alpha1.S3, client client.Client, scheme *runtime.Scheme) error {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.GetName(),
			Namespace: cr.GetNamespace(),
		},
		Spec: v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: "s3.amazonaws.com",
		},
	}

	// TODO: record result in a event
	if _, err := controllerutil.CreateOrUpdate(context.TODO(), client, svc, func() error {
		return controllerutil.SetControllerReference(cr, svc, scheme)
	}); err != nil {
		return err
	}

	return nil
}
