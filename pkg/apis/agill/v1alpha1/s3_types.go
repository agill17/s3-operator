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
	Region string `json:"region,required"`
	BucketName string `json:"bucketName,required"`
	BucketACL string `json:"bucketACL"`
	EnableObjectLock bool `json:"enableObjectLock"`
}

type IAMUser struct {
	Username string `json:"username"`
}


// S3Status defines the observed state of S3
type S3Status struct {
	AccessKeyId string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// S3 is the Schema for the s3s API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=s3s,scope=Namespaced
type S3 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec   S3Spec   `json:"spec,omitempty"`
	Status S3Status `json:"status,omitempty"`
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
