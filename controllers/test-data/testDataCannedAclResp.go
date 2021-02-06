package test_data

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	ExpectedRespBasedOnCannedAclType = map[string]interface{}{
		"log-delivery-write": &s3.GetBucketAclOutput{
			Grants: []*s3.Grant{
				{
					Grantee: &s3.Grantee{
						ID:   aws.String("75aa57f09aa0c8caeab4f8c24e99d10f8e7faeebf76c078efc7c6caea54ba06a"),
						Type: aws.String(s3.TypeCanonicalUser),
					},
					Permission: aws.String(s3.BucketLogsPermissionFullControl),
				},
				{
					Grantee: &s3.Grantee{
						Type: aws.String(s3.TypeGroup),
						URI:  aws.String("http://acs.amazonaws.com/groups/s3/LogDelivery"),
					},
					Permission: aws.String(s3.BucketLogsPermissionWrite),
				},
			},
			Owner: &s3.Owner{
				DisplayName: aws.String("webfile"),
				ID:          aws.String("75aa57f09aa0c8caeab4f8c24e99d10f8e7faeebf76c078efc7c6caea54ba06a"),
			},
		},
	}
)
