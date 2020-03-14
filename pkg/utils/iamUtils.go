package utils

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

func IAMUserExists(username string, iamClient iamiface.IAMAPI) (bool, error) {
	_, err := iamClient.GetUser(&iam.GetUserInput{UserName:&username})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == iam.ErrCodeNoSuchEntityException {
				return false, nil
			}
		}
		// if we are not catching a err, we return false with err
		return false, err
	}
	return true, nil
}

func IAMPolicyExists(policyArn string, iamClient iamiface.IAMAPI) (bool, error) {
	_, err := iamClient.GetPolicy(&iam.GetPolicyInput{PolicyArn:&policyArn})
	if err != nil {
		return false, err
	}
	return true, nil
}

func DeleteAccessKey(accessKey, username string, iamapi iamiface.IAMAPI) error {
	_, err := iamapi.DeleteAccessKey(&iam.DeleteAccessKeyInput{
		AccessKeyId: &accessKey,
		UserName:    &username,
	})
	return err
}

func DeleteAllAccessKeys(iamUser string, iamapi iamiface.IAMAPI) error {
	accessKeys, err := iamapi.ListAccessKeys(&iam.ListAccessKeysInput{
		UserName: &iamUser,
	})
	if err != nil {
		return err
	}

	for _, e := range accessKeys.AccessKeyMetadata {
		if err := DeleteAccessKey(*e.AccessKeyId, iamUser, iamapi) ; err != nil {
			return err
		}
	}
	return nil
}

func GetAccessKeyForUser(username string, iamclient iamiface.IAMAPI) (string, error) {
	keys, err := iamclient.ListAccessKeys(&iam.ListAccessKeysInput{
		UserName: &username,
	})
	if err != nil {
		return "", err
	}

	// TODO: make this better as its bad and risky
	if len(keys.AccessKeyMetadata) == 0 || len(keys.AccessKeyMetadata) > 1 {
		return "", errors.New("IAMUserDoesNotHaveOnly1AccessKeyID")
	}

	return *keys.AccessKeyMetadata[0].AccessKeyId, nil
}

func CreateAccessKeys(username string, iamapi iamiface.IAMAPI) (*iam.CreateAccessKeyOutput, error) {
	return iamapi.CreateAccessKey(&iam.CreateAccessKeyInput{UserName:&username})
}

func CreateIAMUser(input *iam.CreateUserInput, iamClient iamiface.IAMAPI) error {
	userExists, checkErr := IAMUserExists(*input.UserName, iamClient)
	if checkErr != nil {
		return checkErr
	}
	if !userExists {
		if _, errCreatingUser := iamClient.CreateUser(input); errCreatingUser!= nil {
			return errCreatingUser
		}
	}
	return nil
}

func AttachPolicyToIAMUser(username, policyArn string, iamClient iamiface.IAMAPI) error {
	_, err := iamClient.AttachUserPolicy(&iam.AttachUserPolicyInput{
		PolicyArn: &policyArn,
		UserName:  &username,
	})
	return err
}

func DeleteUser(username, policyArn string, iamClient iamiface.IAMAPI) error {

	if errDeletingAccessKeys := DeleteAllAccessKeys(username, iamClient); errDeletingAccessKeys != nil {
		return errDeletingAccessKeys
	}

	if _, errDetachingAccessPolicy := iamClient.DetachUserPolicy(
		&iam.DetachUserPolicyInput{
			UserName:&username,
			PolicyArn:&policyArn},); errDetachingAccessPolicy != nil {
		return errDetachingAccessPolicy
	}

	if _, errDeletingUser := iamClient.DeleteUser(&iam.DeleteUserInput{UserName:&username}); errDeletingUser != nil {
		return errDeletingUser
	}

	return nil
}