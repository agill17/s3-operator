package controllers

import (
	"context"
	"fmt"
	"github.com/agill17/s3-operator/api/v1alpha1"
	"github.com/agill17/s3-operator/controllers/factory"
	test_data "github.com/agill17/s3-operator/controllers/test-data"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"path/filepath"
	"time"
)

const validTestDataFiles = "./test-data/valid"

var _ = Describe("Successful e2e create and delete", func() {

	// TODO: convert these to test form like data instead of files
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
				errCreating := k8sClient.Create(context.TODO(), cr)
				Expect(errCreating).To(BeNil())
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
			It("should match transfer acceleration configuration in AWS", func() {
				By("verifying bucket transfer acceleration in AWS", func() {
					checkBucketTransferAccel(cr)
				})
			})
			It("should match tags configuration in AWS", func() {
				By("verifying bucket tags in AWS", func() {
					checkTags(cr)
				})
			})

			It("should match the canned ACL in AWS", func() {
				By("verifying the canned acl configuration in AWS", func() {
					checkCannedAcl(cr)
				})
			})

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

func checkCannedAcl(cr *v1alpha1.Bucket) {
	out, err := mockS3Client.GetBucketAcl(&s3.GetBucketAclInput{Bucket: aws.String(cr.Spec.BucketName)})
	Expect(err).To(BeNil())
	if cr.Spec.CannedBucketAcl != "" {
		expectedResp, foundInTestData := test_data.ExpectedRespBasedOnCannedAclType[cr.Spec.CannedBucketAcl]
		if !foundInTestData {
			fmt.Printf("************\nWARN: Expected test data for canned "+
				"acl type: %v not found in test_data.ExpectedRespBasedOnCannedAclType"+
				"\n******************\n", cr.Spec.CannedBucketAcl)
			return
		}
		expectedRespGetBucketAclObj, _ := expectedResp.(*s3.GetBucketAclOutput)
		Expect(*out).To(BeEquivalentTo(*expectedRespGetBucketAclObj))

	}
}

var _ = Describe("Negative tests", func() {
	When("A required input is not provided", func() {
		testCrs := []*v1alpha1.Bucket{
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "missing-region",
				},
				Spec: v1alpha1.BucketSpec{
					BucketName: "missing-region",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "missing-bucketName",
				},
				Spec: v1alpha1.BucketSpec{
					Region: "us-east-1",
				},
			},
		}
		for _, test := range testCrs {
			It("Should error out when creating CR", func() {
				Expect(k8sClient.Create(context.TODO(), test)).ShouldNot(Succeed())
			})
			It("Should not exist in cluster", func() {
				crFromCluster := &v1alpha1.Bucket{}
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
					Name:      test.GetName(),
					Namespace: test.GetNamespace(),
				}, crFromCluster)).ShouldNot(Succeed())
			})
			It("Should not exist in AWS", func() {
				_, err := mockS3Client.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: aws.String(test.Spec.BucketName)})
				if err != nil {
					if awsErr, isAwsErr := err.(awserr.Error); isAwsErr {
						Expect(awsErr.Code()).To(BeIdenticalTo(request.InvalidParameterErrCode))
					}
				}
			})
		}
	})

})

func checkBucketTransferAccel(cr *v1alpha1.Bucket) {
	accelOut, err := mockS3Client.GetBucketAccelerateConfiguration(&s3.GetBucketAccelerateConfigurationInput{
		Bucket: aws.String(cr.Spec.BucketName)})
	Expect(err).To(BeNil())

	if !cr.Spec.EnableTransferAcceleration {
		Expect(accelOut.Status).To(BeNil())
		return
	}

	Expect(accelOut.Status).ToNot(BeNil())
	expectedStatus := s3.BucketAccelerateStatusEnabled
	actualStatus := *accelOut.Status
	Expect(actualStatus).To(BeIdenticalTo(expectedStatus))
}

func checkTags(cr *v1alpha1.Bucket) {
	out, err := mockS3Client.GetBucketTagging(&s3.GetBucketTaggingInput{
		Bucket: aws.String(cr.Spec.BucketName),
	})
	if len(cr.Spec.Tags) == 0 {
		Expect(err).ToNot(BeNil())
		return
	}

	expectedMap := cr.Spec.Tags
	actualMap := factory.TSetToMap(out.TagSet)
	Expect(expectedMap).To(BeEquivalentTo(actualMap))
}
