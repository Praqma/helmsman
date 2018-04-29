---
version: v1.2.0-rc
---

# use private helm charts

Helmsman allows you to use private charts from private repos. Currently only repos hosted in S3 or GCS buckets are supported for private repos. 

Other hosting options might be supported in the future. Please open an issue if you require supporting other options.

define your private repo: 

```toml
...

[helmRepos]
stable = "https://kubernetes-charts.storage.googleapis.com"
incubator = "http://storage.googleapis.com/kubernetes-charts-incubator"
myPrivateRepo = s3://this-is-a-private-repo/charts

...

``` 

## S3

If you are using S3 private repos, you need to provide the following AWS env variables:

- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- AWS_DEFAULT_REGION

Helmsman uses the [helm s3](https://github.com/hypnoglow/helm-s3) plugin to work with S3 helm repos.

## GCS

If you are using GCS private repos, you need to provide one of the following env variables:

- `GOOGLE_APPLICATION_CREDENTIALS` environment variable to contain the absolute path to your Google cloud credentials.json file.
- Or, `GCLOUD_CREDENTIALS` environment variable to contain the content of the credentials.json file. 

Helmsman uses the [helm GCS](https://github.com/nouney/helm-gcs) plugin to work with GCS helm repos.