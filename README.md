# S3 Operator

### Overview
A kubernetes operator to create and manage AWS S3 buckets from single control plane. The operator needs AWS access credentials
to manage other resources. It needs permissions to IAM and S3. When a CR is created, the operator will create IAM user
with a inline policy to allow all S3 actions ONLY on the specified S3 bucket ( this way the IAM user being created only has 
access to perform actions against the bucket that gets created). The operator also generates and manages AWS creds for that 
IAM user in a kubernetes secret within the same namespace. If the creds no longer match with the one on AWS, the secret will
get updated with correct access keys. Users also have ability to get new fresh pair of Access keys by simply deleting
the kubernetes secret created by the operator. Finally S3 bucket is created and kept up to date. If the CR is deleted, all 
the IAM resources and S3 resources will get deleted with it.

### Installation
-  Helm Chart ( Helm v3 )
```
  kubectl create ns s3-operator
```
```
  helm upgrade s3-operator \
        chart/s3-operator \
        --install --force --namespace=s3-operator \
        --set AWS_ACCESS_KEY_ID=<YOUR_ACCESS_KEY> \
        --set AWS_SECRET_ACCESS_KEY=<YOUR_SECRET_ACCESS_KEY>
```
- Sample S3 CR can be found [here](https://github.com/agill17/s3-operator/blob/master/deploy/crds/agill.apps_v1alpha1_s3_cr.yaml)

### Features
- [x] Create/Recreate(if deleted from AWS) Bucket
- [x] Delete bucket
- [ ] Update bucket ( need to add more bucket spec to support updates )
- [ ] Bucket spec (properties, management, permissions, etc)
- [x] Restrict IAM user being created to only have permission to s3 bucket
- [x] Create/Recreate/Update kubernetes service of type externalName ( pointing to `s3.amazonaws.com` )
- [x] Create/Recreate (if deleted from AWS) IAM user
- [x] Delete IAM user 
- [x] Create/Recreate/Update secret which contains IAM user credentials ( `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`)
- [x] Get new fresh pair of access keys by deleting kubernetes secret


### Notes
- In addition to event based trigger to reconcile, a periodic sync is also in place to reconcile every n seconds.
    - Default periodic sync period is set to 300 seconds.
    - Can be changed by update `syncPeriod` env variable in operator deployment.

### TODO
- ~~Remove `status.phase` as a way to track reconcile, it adds un-necessary and redundant checks whether a reconcile is needed~~
- Add tags to cloud resources ( as a way to own them ), this way if user or bucket already exists and does not have tags, operator should complain and not perform any actions on it.
