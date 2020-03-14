# S3 Operator

### Overview
### Installation


### Features
- [x] Create/Recreate(if deleted from AWS) Bucket
- [x] Delete bucket
- [ ] Update bucket ( need to add more bucket spec to support updates )
- [ ] Bucket spec (properties, management, permissions, etc)
- [x] Restrict bucket access to IAM user being created (using inline iam policy and attaching to user)
- [x] Create/Recreate/Update kubernetes service of type externalName ( pointing to `s3.amazonaws.com` )
- [x] Create/Recreate (if deleted from AWS) IAM user
- [x] Delete IAM user 
- [x] Create/Recreate/Update secret which contains IAM user credentials ( `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`)
- [x] Get new fresh pair of access keys by deleting kubernetes secret
