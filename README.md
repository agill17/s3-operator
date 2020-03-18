# S3 Operator

### Overview

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
