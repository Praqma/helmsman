---
version: v0.1.2
---

You can manage Helm charts deployment on a hosted K8S cluster in the cloud or on-prem. You need to include the required information to connect to the cluster in your state file. Below is an example:

**IMPORTANT**: Only Certificates and private helm repos in S3 buckets are currently supported. Helmsman needs valid AWS access keys to be able to retrieve private charts or certificates from your s3 buckets. It expects the keys to be in the following environemnt variables:

- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- AWS_DEFAULT_REGION

Also, the K8S user password is expected in an environment variable which you can give any name you want and define it in your desired state file.

```
[metadata]
org = "orgX"
maintainer = "k8s-admin"

# Certificates are used to connect to the cluster. Currently, they can only be retrieved from s3 buckets.
[certificates]
caCrt = "s3://your-bucket/ca.crt" 
caKey = "s3://your-bucket/ca.key" 

[settings]
kubeContext = "mycontext" 
username = "<<your-username>>"
password = "$PASSWORD" # the name of an environment variable containing the k8s password
clusterURI = "<<your_cluster_API_URI>>" # cluster API

[namespaces]
staging = "staging" 

[helmRepos]
stable = "https://kubernetes-charts.storage.googleapis.com"
incubator = "http://storage.googleapis.com/kubernetes-charts-incubator"
myrepo = "s3://my-private-repo/charts"

[apps]

    [apps.jenkins]
    name = "jenkins" 
    description = "jenkins"
    env = "staging" 
    enabled = true 
    chart = "stable/jenkins" 
    version = "0.9.1" 
    valuesFile = "" 
    purge = false 
    test = false 


    [apps.artifactory]
    name = "artifactory" 
    description = "artifactory"
    env = "staging" 
    enabled = true 
    chart = "stable/artifactory" 
    version = "6.2.0" 
    valuesFile = "" 
    purge = false 
    test = false 
```