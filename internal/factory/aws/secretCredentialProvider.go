package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

type k8sSecretCredentialProvider struct {
	secretData map[string][]byte
}

func NewK8sSecretCredentialProvider(sData map[string][]byte) *credentials.Credentials {
	return credentials.NewCredentials(&k8sSecretCredentialProvider{secretData: sData})
}
// Retrieve returns nil if it successfully retrieved the value.
// Error is returned if the value were not obtainable, or empty.
func (a *k8sSecretCredentialProvider) Retrieve() (credentials.Value, error) {
	aKeyID, hasAKeyId := a.secretData[awsAccessKeyID]
	sAccessKey, hasSAccessKey := a.secretData[awsSecretAccessKey]
	if hasAKeyId && hasSAccessKey {
		return credentials.Value{
			AccessKeyID:     string(aKeyID),
			SecretAccessKey: string(sAccessKey),
			SessionToken:    string(a.secretData["AWS_SESSION_TOKEN"]),
			ProviderName:    "K8sSecretData",
		}, nil
	}
	errMsg := fmt.Sprintf("ErrAWSFailedToExtractCredentialsFromK8sSecret: Found accessKeyID: %t, Found secretAccessKey: %t", hasAKeyId, hasSAccessKey)
	return credentials.Value{}, &ErrAWSFailedToExtractCredentialsFromK8sSecret{Message: errMsg}
}

// IsExpired returns if the credentials are no longer valid, and need
// to be retrieved.
func (a *k8sSecretCredentialProvider) IsExpired() bool {
	return false
}
