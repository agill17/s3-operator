package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// S3Spec defines the desired state of S3
type S3Spec struct {
	// +optional
	IAMUserSpec IAMUser `json:"iamUser"`

	// +kubebuilder:validation:Required
	Region string `json:"region,required"`

	// +kubebuilder:validation:Required
	BucketName string `json:"bucketName,required"`

	// The canned ACL to apply to the bucket.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=private;public-read;public-read-write;authenticated-read
	BucketACL string `json:"bucketACL,required"`

	// Specifies whether you want S3 Object Lock to be enabled for the new bucket.
	// +optional
	EnableObjectLock bool `json:"enableObjectLock,omitempty"`

	// Decides whether versioning should be enabled. Defaults to false.
	// +optional
	EnableVersioning bool `json:"enableVersioning,omitempty"`

	// Decides whether transfer acceleration should be enabled. Defaults to false
	// +optional
	EnableTransferAcceleration bool `json:"enableTransferAcceleration,omitempty"`

	// +optional
	BucketPolicy string `json:"bucketPolicy,omitempty"`
}

type IAMUser struct {
	Username string `json:"username"`
}

// S3Status defines the observed state of S3
type S3Status struct {
	Status string `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// S3 is the Schema for the s3s API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=s3s,scope=Namespaced
// +kubebuilder:printcolumn:name="bucket-name",type=string,JSONPath=`.spec.bucketName`
// +kubebuilder:printcolumn:name="IAM-User",type=string,JSONPath=`.spec.iamUser.username`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
type S3 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              S3Spec   `json:"spec,omitempty"`
	Status            S3Status `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// S3List contains a list of S3
type S3List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S3 `json:"items"`
}

func init() {
	SchemeBuilder.Register(&S3{}, &S3List{})
}
