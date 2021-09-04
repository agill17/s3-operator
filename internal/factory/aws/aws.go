package aws

import (
	"context"
	"github.com/agill17/s3-operator/api/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func (a *awsClient) BucketExists(in *v1alpha1.Bucket) (bool, error) {
	_, err := a.s3Client.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: aws.String(in.Spec.BucketName)})
	if err != nil {
		if awsErr, isAwsErr := err.(awserr.Error); isAwsErr && awsErr.Code() == s3.ErrCodeNoSuchBucket {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (a *awsClient) CreateBucket(in *v1alpha1.Bucket) error {
	_, err := a.s3Client.CreateBucket(in.AWSCreateBucketIn())
	return err
}

func (a *awsClient) DeleteBucket(in *v1alpha1.Bucket) error {
	exists, err := a.BucketExists(in)
	if err != nil {
		return err
	}

	if exists {
		iter := s3manager.NewDeleteListIterator(a.s3Client, &s3.ListObjectsInput{
			Bucket: aws.String(in.Spec.BucketName),
		})

		// Traverse iterator deleting each object
		err = s3manager.NewBatchDeleteWithClient(a.s3Client).Delete(context.TODO(), iter)
		if err != nil {
			return err
		}

		_, err = a.s3Client.DeleteBucket(in.AWSDeleteBucketIn())
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *awsClient) ApplyBucketProperties(in *v1alpha1.Bucket) error {

	bucketVersioningInput := in.AWSPutBucketVersioningIn()
	_, errApplyingVersioning := a.s3Client.PutBucketVersioning(bucketVersioningInput)
	if errApplyingVersioning != nil {
		return errApplyingVersioning
	}

	bucketAccelConfigIn := in.AWSBucketAccelerationConfigIn()
	_, errApplyingTACL := a.s3Client.PutBucketAccelerateConfiguration(bucketAccelConfigIn)
	if errApplyingTACL != nil {
		return errApplyingTACL
	}

	policyIn := in.AWSBucketPolicyInput()
	errFailedValidation := policyIn.Validate()
	if errFailedValidation != nil {
		return errFailedValidation
	}
	_, err := a.s3Client.PutBucketPolicy(policyIn)
	if err != nil {
		return err
	}

	_, err = a.s3Client.PutBucketAcl(in.AWSPutBucketCannedAclInput())
	if err != nil {
		return err
	}

	_, err = a.s3Client.PutBucketTagging(in.AWSPutTagsIn(MapToTagging(in.Spec.Tags)))
	if err != nil {
		return err
	}

	return nil
}

func MapToTagging(m map[string]string) *s3.Tagging {
	t := &s3.Tagging{TagSet: []*s3.Tag{}}
	for k, v := range m {
		t.TagSet = append(t.TagSet, &s3.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	return t
}

func TagListToTagMap(tSet []*s3.Tag) map[string]string {
	m := map[string]string{}
	for _, ele := range tSet {
		m[*ele.Key] = *ele.Value
	}
	return m

}
