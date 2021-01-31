package controllers

import (
	"context"
	"fmt"
	"github.com/agill17/s3-operator/api/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"path/filepath"
	"time"
)

const validTestDataFiles = "./test-data/valid"
const enabledStatus = "Enabled"
const disabledStatus = "Suspended"

var _ = Describe("Successful e2e create and delete", func() {

	// test-data prep
	tests, errReadingDir := ioutil.ReadDir(validTestDataFiles)
	if errReadingDir != nil {
		Fail("Failed to read valid test dir")
	}

	for _, test := range tests {

		rawTestData, errReading := ioutil.ReadFile(filepath.Join(validTestDataFiles, test.Name()))
		if errReading != nil {
			Fail("Failed to read valid test file")
		}

		cr := &v1alpha1.Bucket{}
		err := yaml.Unmarshal(rawTestData, cr)
		if err != nil {
			Fail("Failed to parse test file")
		}
		namespacedName := fmt.Sprintf("%s/%s", cr.GetNamespace(), cr.GetName())

		// create and verify in cluster and verify in AWS
		When(fmt.Sprintf("%v: bucket cr is applied in cluster", namespacedName), func() {
			It("Gets created and reconciled successfully", func() {
				Expect(k8sClient.Create(context.TODO(), cr)).Should(Succeed())
				bucketCrFromCluster := &v1alpha1.Bucket{}
				By(fmt.Sprintf("Verifying that %v CR exists and status is created", namespacedName), func() {

					Eventually(func() bool {
						if err := k8sClient.Get(context.TODO(), types.NamespacedName{
							Name:      cr.GetName(),
							Namespace: cr.GetNamespace()},
							bucketCrFromCluster); err != nil {
							return false
						}
						return bucketCrFromCluster.Status.Ready

					}, 30*time.Second, 2*time.Second).Should(BeTrue())
				})
			})
		})

		When(fmt.Sprintf("bucket %v is created successfully in cluster", namespacedName), func() {
			It("should exist in AWS", func() {
				By(fmt.Sprintf("Verifying that %v bucket is created in AWS", namespacedName), func() {
					Eventually(func() error {
						_, err := mockS3Client.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: aws.String(cr.Spec.BucketName)})
						return err
					}, 10*time.Second, 2*time.Second).Should(BeNil())

				})
			})
		})

		When(fmt.Sprintf("bucket %v is created successfully in AWS", namespacedName), func() {
			It("should match versioning configuration in AWS", func() {
				By("verifying bucket versioning in AWS", func() {
					checkBucketVersioning(cr)
				})
			})
			//It("should match transfer acceleration configuration in AWS", func() {
			//	By("verifying bucket transfer acceleration in AWS", func() {
			//		checkBucketTransferAccel(cr)
			//	})
			//})
		})

		// delete and verify in cluster and verify in AWS
		When(fmt.Sprintf("bucket %v cr is deleted from cluster", namespacedName), func() {
			It("CR should no longer exists in cluster", func() {
				Expect(k8sClient.Delete(context.TODO(), cr)).Should(Succeed())

				By(fmt.Sprintf("verifying %v cr no longer exists in cluster", namespacedName), func() {
					Eventually(func() bool {
						clusterCr := &v1alpha1.Bucket{}
						err := k8sClient.Get(context.TODO(), types.NamespacedName{
							Name:      cr.GetName(),
							Namespace: cr.GetNamespace(),
						}, clusterCr)
						if err != nil {
							if errors.IsNotFound(err) {
								return true
							}
						}
						return false
					}, 15*time.Second, 5*time.Second).Should(BeTrue())

				})
			})
		})

		When("Bucket CR is deleted from cluster", func() {
			It(fmt.Sprintf("bucket %v should no longer exist in AWS", namespacedName), func() {
				By(fmt.Sprintf("verifing %v bucket does not exist in AWS", namespacedName), func() {
					Eventually(func() bool {
						_, err := mockS3Client.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: aws.String(cr.Spec.BucketName)})
						if err != nil {
							if awsErr, isAwsErr := err.(awserr.Error); isAwsErr && awsErr.Code() == s3.ErrCodeNoSuchBucket {
								return true
							}
						}
						return false
					}, 5*time.Second, 2*time.Second).Should(BeTrue())
				})
			})
		})
	}
})

func checkBucketVersioning(cr *v1alpha1.Bucket) {
	out, err := mockS3Client.GetBucketVersioning(&s3.GetBucketVersioningInput{Bucket: aws.String(cr.Spec.BucketName)})
	Expect(err).To(BeNil())
	Expect(out.Status).ToNot(BeNil())
	expectedStatus := s3.BucketVersioningStatusSuspended
	if cr.Spec.EnableVersioning {
		expectedStatus = s3.BucketVersioningStatusEnabled
	}
	actualStatus := *out.Status
	Expect(actualStatus).To(BeIdenticalTo(expectedStatus))

}

func checkBucketTransferAccel(cr *v1alpha1.Bucket) {
	accelOut, err := mockS3Client.GetBucketAccelerateConfiguration(&s3.GetBucketAccelerateConfigurationInput{
		Bucket: aws.String(cr.Spec.BucketName)})
	Expect(err).To(BeNil())
	Expect(accelOut.Status).ToNot(BeNil())

	expectedStatus := s3.BucketAccelerateStatusSuspended
	if cr.Spec.EnableTransferAcceleration {
		expectedStatus = s3.BucketAccelerateStatusEnabled
	}
	actualStatus := *accelOut.Status
	Expect(actualStatus).To(BeIdenticalTo(expectedStatus))

}

//TODO: add negative cases
/**
1. required field not passed in
2. bad input ( failure on aws side )
*/
