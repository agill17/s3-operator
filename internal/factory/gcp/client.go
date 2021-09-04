package gcp

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	serviceAccountCredJsonSecretKey = "GOOGLE_CLOUD_CREDENTIALS"
)

type gcpClient struct {
	storageClient *storage.Client
	projectID string
}

func NewGCPBucket(c client.Client, sRef *v1.SecretReference) (*gcpClient, error){
	creds, err := getCredentials(c, sRef)
	if err != nil {
		return nil, err
	}
	client, err := storage.NewClient(context.TODO(), option.WithCredentials(creds))
	if err != nil {
		return nil, err
	}
	return &gcpClient{storageClient: client, projectID: creds.ProjectID}, nil
}


func getCredentials(c client.Client, sRef *v1.SecretReference) (*google.Credentials, error) {
	if sRef == nil {
		return google.FindDefaultCredentials(context.TODO())
	}
	secret := &v1.Secret{}
	err := c.Get(context.TODO(),
		types.NamespacedName{
			Name: sRef.Name,
			Namespace: sRef.Namespace,
		}, secret)
	if err != nil {
		return nil, err
	}
	saJsonData, ok := secret.Data[serviceAccountCredJsonSecretKey]
	if !ok {
		errMsg := fmt.Sprintf("Could not find %q in %q/%q secret", serviceAccountCredJsonSecretKey,
			sRef.Namespace, sRef.Name)
		return nil, &ErrGCPFailedToExtractCredentialsFromK8sSecret{Message: errMsg}
	}
	return google.CredentialsFromJSON(context.TODO(),saJsonData)

}