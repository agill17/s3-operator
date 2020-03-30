package s3

import (
	"context"
	"github.com/agill17/s3-operator/pkg/utils"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func DeleteUser(username string, iamClient iamiface.IAMAPI) error {
	userExists, err := utils.IAMUserExists(username, iamClient)
	if err != nil {
		return err
	}

	if userExists {
		if errDeletingAccessKeys := utils.DeleteAllAccessKeys(username, iamClient); errDeletingAccessKeys != nil {
			return errDeletingAccessKeys
		}

		if errDetachingPolicy := utils.DeleteAllUserInlinePolicies(username, iamClient); errDetachingPolicy != nil {
			return errDetachingPolicy
		}

		if _, errDeletingUser := iamClient.DeleteUser(&iam.DeleteUserInput{UserName: &username}); errDeletingUser != nil {
			return errDeletingUser
		}
	}

	return nil
}

// Assumes empty the bucket , then delete
func DeleteBucket(bucketName string, s3Client s3iface.S3API) error {

	exists, err := utils.BucketExists(bucketName, s3Client)
	if err != nil {
		return err
	}

	if exists {
		iter := s3manager.NewDeleteListIterator(s3Client, &s3.ListObjectsInput{
			Bucket: &bucketName,
		})
		// Traverse iterator deleting each object
		if err := s3manager.NewBatchDeleteWithClient(s3Client).Delete(context.TODO(), iter); err != nil {
			return err
		}

		if _, errDeleting := s3Client.DeleteBucket(&s3.DeleteBucketInput{Bucket: &bucketName}); errDeleting != nil {
			return errDeleting
		}
	}
	return nil
}
