package utils

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

func BucketExists(bucketName string, s3Client s3iface.S3API) (bool, error) {
	_, err := s3Client.GetBucketLocation(&s3.GetBucketLocationInput{Bucket:&bucketName})
	if err != nil {
		return false, err
	}
	return true, nil
}

// TODO: Think about reconciling bucket ACL
//func GetBucketACL(bucketName string, s3Client s3iface.S3API) error {
//	bucketAcl, err := s3Client.GetBucketAcl(&s3.GetBucketAclInput{Bucket:aws.String(bucketName)})
//	if err != nil {
//		return err
//	}
//
//}
