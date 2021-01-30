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
			It("Should set the cr status to created", func() {
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

			It(fmt.Sprintf("%v: bucket should exist in AWS", namespacedName), func() {
				By(fmt.Sprintf("Verifying that %v bucket is created in AWS", namespacedName), func() {

					Eventually(func() error {
						_, err := mockS3Client.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: aws.String(cr.Spec.BucketName)})
						return err
					}, 10*time.Second, 2*time.Second).Should(BeNil())

				})
			})

			// TODO: not yet implemented in controller
			if cr.Spec.EnableVersioning {
				It("bucket versioning should be enabled in AWS", func() {
					By("verifying versioning is enabled in AWS bucket", func() {

						Eventually(func() bool {
							out, err := mockS3Client.GetBucketVersioning(&s3.GetBucketVersioningInput{
								Bucket: aws.String(cr.Spec.BucketName),
							})
							if err != nil {
								return false
							}
							return aws.StringValue(out.Status) == "true"

						}).Should(BeTrue())

					})
				})
			}
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
					}, 10*time.Second, 2*time.Second).Should(BeTrue())

				})
			})
		})
	}
})

//TODO: add negative cases
/**
1. required field not passed in
2. bad input ( failure on aws side )
*/
