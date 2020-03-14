# S3 Operator

### Overview
### Installation


### Features
- [x] Create/Recreate(if deleted from AWS) Bucket
- [ ] Delete bucket ( coming soon ) 
- [ ] Update bucket
- [ ] Bucket spec (properties, management, permissions, etc)
- [x] Create/Recreate/Update kubernetes service of type externalName ( pointing to `s3.amazonaws.com` )
- [x] Create/Recreate (if deleted from AWS) IAM user
- [ ] Delete IAM user ( coming soon ) 
- [x] Create/Recreate/Update secret which contains IAM user credentials ( `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`)
- [x] Get new fresh pair of access keys by deleting kubernetes secret
