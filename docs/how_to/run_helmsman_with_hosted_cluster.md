---
version: v0.2.0
---

You can manage Helm charts deployment on a hosted K8S cluster in the cloud or on-prem. You need to include the required information to connect to the cluster in your state file. 

**IMPORTANT**: Helmsman expects certain environment variables to be available depending on where your cluster and connection certificates are hosted. Certificates can be used from S3/GCS buckets or local file system. 

## AWS
If you use s3 buckets for storing certificates or for hosting private helm repos, Helmsman needs valid AWS access keys to be able to retrieve private charts or certificates from your s3 buckets. It expects the keys to be in the following environemnt variables:

- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- AWS_DEFAULT_REGION

## GCS
If you use GCS buckets for storing certificates or for hosting private helm repos, Helmsman needs valid Google Cloud credentials to authenticate reading requests from private buckets. This can be provided in one of two ways: 

- set `GOOGLE_APPLICATION_CREDENTIALS` environment variable to contain the absolute path to your Google cloud credentials.json file.
- Or, set `GCLOUD_CREDENTIALS` environment variable to contain the content of the credentials.json file. 

check [here](https://www.terraform.io/docs/providers/google/index.html#authentication-json-file) for getting the required authentication file.

## Additional environemnt variables

The K8S user password is expected in an environment variable which you can give any name you want and define it in your desired state file. Additionally, you can optionally use environment variables to provide certificate paths and clusterURI.


Below is an example state file:

```
[metadata]
org = "orgX"
maintainer = "k8s-admin"

# Certificates are used to connect to the cluster. Currently, they can only be retrieved from s3 buckets.
[certificates]
caCrt = "s3://your-bucket/ca.crt" # s3 bucket
caKey = "$K8S_CLIENT_KEY" # relative file path
caClient = "gs://your-GCS-bucket/caClient.crt" # GCS bucket

[settings]
kubeContext = "mycontext" 
username = "<<your-username>>"
password = "$K8S_PASSWORD" # the name of an environment variable containing the k8s password
clusterURI = "$K8S_URI" # cluster API

[namespaces]
staging = "staging" 

[helmRepos]
stable = "https://kubernetes-charts.storage.googleapis.com"
incubator = "http://storage.googleapis.com/kubernetes-charts-incubator"
myrepo = "s3://my-private-repo/charts"
myGCSrepo = "gs://my-GCS-repo/charts"

[apps]

    [apps.jenkins]
    name = "jenkins" 
    description = "jenkins"
    namespace = "staging" 
    enabled = true 
    chart = "stable/jenkins" 
    version = "0.9.1" 
    valuesFile = "" 
    purge = false 
    test = false 


    [apps.artifactory]
    name = "artifactory" 
    description = "artifactory"
    namespace = "staging" 
    enabled = true 
    chart = "stable/artifactory" 
    version = "6.2.0" 
    valuesFile = "" 
    purge = false 
    test = false 
```

The above example requires the following environment variables to be set:

- AWS_ACCESS_KEY_ID (since S3 is used for helm repo and certificates)
- AWS_SECRET_ACCESS_KEY
- AWS_DEFAULT_REGION
- GOOGLE_APPLICATION_CREDENTIALS (since GCS is used for helm repo and certificates)
- K8S_CLIENT_KEY (used in the file)
- K8S_PASSWORD (used in the file)
- K8S_URI (used in the file)