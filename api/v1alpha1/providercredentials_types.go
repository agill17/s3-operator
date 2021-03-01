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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProviderType string

const (
	ProviderTypeAWS   ProviderType = "aws"
	ProviderTypeGCP   ProviderType = "gcp"
	ProviderTypeAzure ProviderType = "azure"
)

// ProviderSpec defines the desired state of Provider
type ProviderSpec struct {
	// aws, gcp(not-yet-supported), azure(not-yet-supported)
	Type ProviderType `json:"type,required"`

	// +optional
	// Credentials is a key(string) value(base64-encoded-string) pair for specifying provider specific credentials.
	// We look for default key names per provider. For example in case of aws, we look for AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY
	// The values can be:
	// 1. base64 encoded actual credentials
	// 2. base64 encoded vault paths that have credentials.
	// 	- When using vault paths, the following keys are required in spec.credentials
	// 		1. VAULT_K8S_AUTH_BACKEND_PATH - k8s auth backend path ( example: auth/kubernetes )
	// 		2. VAULT_K8S_AUTH_BACKEND_ROLE - k8s auth role name ( example: test )
	// 	- When using vault paths, the following keys are optional
	// 		1. VAULT_ADDR
	// 		2. VAULT_SKIP_VERIFY - skip tls verification
	// 	- Examples:
	// 		1. AWS provider credentials with vault paths as example:
	// 			spec:
	//				credentials:
	//					VAULT_ADDR: aHR0cDovL3ZhdWx0OjgyMDA=
	//					VAULT_SKIP_VERIFY: dHJ1ZQ==
	//					VAULT_K8S_AUTH_BACKEND_PATH: YXV0aC9raW5k
	//					VAULT_K8S_AUTH_BACKEND_ROLE: dGVzdA==
	//					AWS_ACCESS_KEY_ID: dmF1bHQ6c2VjcmV0L2RhdGEva2luZC9zMy1vcGVyYXRvciNBV1NfQUNDRVNTX0tFWV9JRA==
	//					AWS_SECRET_ACCESS_KEY: dmF1bHQ6c2VjcmV0L2RhdGEva2luZC9zMy1vcGVyYXRvciNBV1NfU0VDUkVUX0FDQ0VTU19LRVk=
	//		2. AWS provider credentials without vault:
	//			spec:
	//				credentials:
	//					AWS_ACCESS_KEY_ID: QVdTX0FDQ0VTU19LRVlfSUQK
	//					AWS_SECRET_ACCESS_KEY: QVdTX1NFQ1JFVF9BQ0NFU1NfS0VZCg==
	Credentials map[string][]byte `json:"credentials,omitempty"`
}

// ProviderStatus defines the observed state of Provider
type ProviderStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// Provider is the Schema for the providercredentials API
// TODO: Provider can JUST become "Provider" and it can accept "providerType"(aws,gcp) and "credentials" in future.
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ProviderSpec   `json:"spec,required"`
	Status            ProviderStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProviderList contains a list of Provider
type ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Provider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Provider{}, &ProviderList{})
}
