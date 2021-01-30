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

package controllers

import (
	"github.com/agill17/s3-operator/controllers/factory"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"k8s.io/apimachinery/pkg/runtime"
	"math"
	"os"
	"path/filepath"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	agillappsv1alpha1 "github.com/agill17/s3-operator/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var testRuntimeScheme = runtime.NewScheme()
var mockS3Client s3iface.S3API
func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = agillappsv1alpha1.AddToScheme(testRuntimeScheme)
	Expect(err).NotTo(HaveOccurred())

	err = scheme.AddToScheme(testRuntimeScheme)
	Expect(err).NotTo(HaveOccurred())

	testMgr, errCreatingNewMgr := controllerruntime.NewManager(cfg, controllerruntime.Options{
		Scheme: testRuntimeScheme,
	})
	Expect(errCreatingNewMgr).NotTo(HaveOccurred())

	err = (&BucketReconciler{
		Scheme:   testRuntimeScheme,
		Client:   testMgr.GetClient(),
		Log:      controllerruntime.Log.WithName("test")}).
		SetupWithManager(testMgr)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient = testMgr.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	// ensure env var for mock s3 server is set
	Expect(os.Getenv(factory.EnvVarS3Endpoint)).ToNot(BeEmpty())

	go func() {
		err = testMgr.Start(controllerruntime.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	cfg := &aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
		MaxRetries:                    aws.Int(math.MaxInt64),
		Region: aws.String("us-east-1"),
		Endpoint: aws.String(os.Getenv(factory.EnvVarS3Endpoint)),
		DisableSSL: aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}

	sess := session.Must(session.NewSession())
	mockS3Client = s3.New(sess, cfg)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
