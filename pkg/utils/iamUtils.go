package utils

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"net/url"
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

func DeleteAllUserInlinePolicies(iamUser string, iamapi iamiface.IAMAPI) error {
	//To list the inline policies for a user, use the ListUserPolicies API.
	allPolicies, err := iamapi.ListUserPolicies(&iam.ListUserPoliciesInput{
		UserName: &iamUser,
	})
	if err != nil {
		return err
	}

	// none found..
	if allPolicies == nil {
		return nil
	}

	for _, e := range allPolicies.PolicyNames {
		if errDetaching := DeleteIAMInlinePolicyFromUser(*e, iamUser, iamapi); errDetaching != nil {
			return errDetaching
		}
	}
	return  nil
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


func IAMPolicyMatchesDesiredPolicyDocument(desiredPolicyDocument, username, policyName string, iamClient iamiface.IAMAPI) (bool, error){
	currentPolicyInAWS, errGetting := iamClient.GetUserPolicy(&iam.GetUserPolicyInput{
		PolicyName: &policyName,
		UserName:   &username,
	})
	if errGetting != nil {
		return false, errGetting
	}
	currentPolicyDocInAws, err := url.QueryUnescape(*currentPolicyInAWS.PolicyDocument)
	if err != nil {
		return false, err
	}
	return desiredPolicyDocument == currentPolicyDocInAws, nil

}

func DeleteIAMInlinePolicyFromUser(policyName, username string, iamClient iamiface.IAMAPI) error {
	_, errDeletingInlinePolicy := iamClient.DeleteUserPolicy(&iam.DeleteUserPolicyInput{
		PolicyName: &policyName,
		UserName:   &username,
	})
	return errDeletingInlinePolicy
}


func IAMUserPolicyExists(policyName, iamUser string, iamClient iamiface.IAMAPI) (bool, error) {
	_, err := iamClient.GetUserPolicy(&iam.GetUserPolicyInput{
		PolicyName: &policyName,
		UserName:   &iamUser,
	})
	if err != nil {
		if awsErr, isAwsErr := err.(awserr.Error); isAwsErr {
			if awsErr.Code() == iam.ErrCodeNoSuchEntityException {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}