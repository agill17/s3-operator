module github.com/agill17/s3-operator

go 1.16

require (
	github.com/aws/aws-sdk-go v1.37.1
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.3.0
	github.com/hashicorp/vault/api v1.0.4
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/spf13/cast v1.3.0
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.2
)
