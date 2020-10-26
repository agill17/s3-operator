# S3 Operator

### Current status
Refactoring to support multiple cloud provider. Refactor work can be found [here](https://github.com/agill17/s3-operator/tree/feature/multi-provider).

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
|Features                               | Create | Delete   | Update |
|---------------------------------------|--------|----------|--------|
| S3 Bucket                             | ✅     | ✅     |       |
| Bucket Versioning                     | ✅     | ✅     | ✅    |
| Bucket Transfer Acceleration          | ✅     | ✅     | ✅    |
| Bucket Canned ACL                     | ✅     | ✅     | ✅    |
| Bucket Policy                         | ✅     | ✅     | ✅    |
| S3 Object Locking ( only on create)   | ✅     |        |        |
| Bucket Transfer Acceleration          | ✅     | ✅     | ✅    |
| Kubernetes service for s3             | ✅     | ✅     | ✅    |
| IAM user                              | ✅     | ✅     | ✅    |
| IAM user restricted access to bucket  | ✅     | ✅     | ✅    |
| IAM user access keys                  | ✅     | ✅     | ✅    |
| IAM user access keys in k8s secret    | ✅     | ✅     | ✅    |







### Additional Notes
- Rotate IAM access keys by deleting k8s secret.
- In addition to event based trigger to reconcile, a periodic sync is also in place to reconcile every n seconds.
    - Default periodic sync period is set to 300 seconds.
    - Can be changed by update `syncPeriod` env variable in operator deployment.

### TODO
- Add tags to cloud resources ( as a way to own them ), this way if user or bucket already exists and does not have tags, operator should complain and not perform any actions on it.
- More bucket properties...
