package utils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"math"
)

func S3Client(region string) s3iface.S3API {
	sess, _ := session.NewSession(&aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
		Region: aws.String(region),
		MaxRetries: aws.Int(math.MaxInt64),
	})
	return s3.New(sess)
}


func IAMClient(region string) iamiface.IAMAPI {
	sess, _ := session.NewSession(&aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
		Region: aws.String(region),
		MaxRetries: aws.Int(math.MaxInt64),
	})
	return iam.New(sess)
}

