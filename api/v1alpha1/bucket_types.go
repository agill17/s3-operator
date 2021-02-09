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

// BucketSpec defines the desired state of Bucket
type BucketSpec struct {
	Region     string `json:"region,required"`
	BucketName string `json:"bucketName,required"`
	// +optional
	EnableVersioning bool `json:"enableVersioning,omitempty"`
	// +optional
	EnableObjectLock bool `json:"enableObjectLock,omitempty"`
	// +optional
	EnableTransferAcceleration bool `json:"enableTransferAcceleration,omitempty"`
	// +optional
	BucketPolicy string `json:"bucketPolicy,omitempty"`
	// +optional
	/**
	Amazon S3 supports a set of predefined grants, known as canned ACLs.
	Each canned ACL has a predefined set of grantees and permissions.
	The following link ( https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#specifying-grantee )
	lists the set of canned ACLs and the associated predefined grants.
	*/
	// +kubebuilder:validation:Enum=private;public-read;public-read-write;aws-exec-read;authenticated-read;bucket-owner-read;bucket-owner-full-control;log-delivery-write
	// +kubebuilder:default=private
	CannedBucketAcl string `json:"cannedBucketAcl,omitempty"`
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// BucketStatus defines the observed state of Bucket
type BucketStatus struct {
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// Bucket is the Schema for the buckets API
// +kubebuilder:printcolumn:name="bucket",type=string,JSONPath=`.spec.bucketName`
// +kubebuilder:printcolumn:name="ready",type=string,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="ACL",type=string,JSONPath=`.spec.cannedBucketAcl`
// Bucket is the Schema for the buckets API
type Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketSpec   `json:"spec,omitempty"`
	Status BucketStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BucketList contains a list of Bucket
type BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Bucket{}, &BucketList{})
}
