---
version: v0.1.3
---

# use private helm charts

Helmsman allows you to use private charts from private repos. Currently only repos hosted in S3 buckets are supported for private repos. 
Other hosting options will be supported in the future. 

define your private repo: 

```
...

[helmRepos]
stable = "https://kubernetes-charts.storage.googleapis.com"
incubator = "http://storage.googleapis.com/kubernetes-charts-incubator"
myPrivateRepo = s3://this-is-a-private-repo/charts

...
``` 

If you are using S3 private repos, you need to provide the following AWS env variables:

- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- AWS_DEFAULT_REGION

Helmsman uses the [helm s3](https://github.com/hypnoglow/helm-s3) plugin to work with S3 helm repos.