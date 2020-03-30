package utils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/davecgh/go-spew/spew"
)

func BucketExists(bucketName string, s3Client s3iface.S3API) (bool, error) {
	_, err := s3Client.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: &bucketName})
	if awserr, ok := err.(awserr.Error); ok && awserr.Code() == s3.ErrCodeNoSuchBucket {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func GetBucketACL(bucketName string, s3Client s3iface.S3API) (*s3.GetBucketAclOutput, error) {
	bucketAcl, err := s3Client.GetBucketAcl(&s3.GetBucketAclInput{Bucket: aws.String(bucketName)})
	if err != nil {
		return nil, err
	}
	return bucketAcl, nil
}

func CreateBucket(createIn *s3.CreateBucketInput, s3Client s3iface.S3API) error {
	bucketExists, checkError := BucketExists(*createIn.Bucket, s3Client)
	if checkError != nil {
		return checkError
	}
	if !bucketExists {
		out, errCreatingBucket := s3Client.CreateBucket(createIn)
		if errCreatingBucket != nil {
			return errCreatingBucket
		}
		spew.Dump(out)
	}
	return nil
}
