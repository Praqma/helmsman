---
version: v1.3.0-rc
---

# Helmsman desired state specification

This document describes the specification for how to write your Helm charts desired state file. This can be either toml or yaml file. The desired state file consists of:

- [Metadata](#metadata) [Optional] -- metadata for any human reader of the desired state file.
- [Certificates](#certificates) [Optional] -- only needed when you want Helmsman to connect kubectl to your cluster for you.
- [Settings](#settings) -- data about your k8s cluster. 
- [Namespaces](#namespaces) -- defines the namespaces where you want your Helm charts to be deployed.
- [Helm Repos](#helm-repos) -- defines the repos where you want to get Helm charts from.
- [Apps](#apps) -- defines the applications/charts you want to manage in your cluster.

## Metadata

Optional : Yes.

Synopsis: Metadata is used for the human reader of the desired state file. While it is optional, we recommend having a maintainer and scope/cluster metadata.

Options: 
- you can define any key/value pairs.

Example: 

```toml
[metadata]
scope = "cluster foo"
maintainer = "k8s-admin"
```

```yaml
metadata:
  scope: "cluster foo"
  maintainer: "k8s-admin"
```

## Certificates

Optional : Yes, only needed if you want Helmsman to connect kubectl to your cluster for you.

Synopsis: defines where to find the certifactions needed for connecting kubectl to a k8s cluster. If connection settings (username/password/clusterAPI) are provided in the Settings section below, then you need AT LEAST to provide caCrt and caKey. You can optionally provide a client certificate (caClient) depending on your cluster connection setup.

Options: 
- caCrt : a valid S3/GCS bucket or local relative file path to a certificate file. 
- caKey : a valid S3/GCS bucket or local relative file path to a client key file.
- caClient: a valid S3/GCS bucket or local relative file path to a client certificate file.

> You can use environment variables to pass the values of the options above. The environment variable name should start with $

> bucket format is: <s3 or gs>://bucket-name/dir1/dir2/.../file.extension

Example: 

```toml
[certificates]
caCrt = "s3://myS3bucket/mydir/ca.crt" 
caKey = "gs://myGCSbucket/ca.key" 
caClient ="../path/to/my/local/client-certificate.crt"
#caClient = "$CA_CLIENT"
```

```yaml
certificates:
  caCrt: "s3://myS3bucket/mydir/ca.crt"
  caKey: "gs://myGCSbucket/ca.key"
  caClient: "../path/to/my/local/client-certificate.crt"
  #caClient: "$CA_CLIENT"
```

## Settings

Optional : No.

Synopsis: provides settings for connecting to your k8s cluster and configuring Helm's Tiller in the cluster.

Options: 
- kubeContext : this is always required and defines what context to use in kubectl. Helmsman will try connect to this context first, if it does not exist, it will try to create it (i.e. connect to a k8s cluster) using the options below.

The following options can be skipped if your kubectl context is already created and you don't want Helmsman to connect kubectl to your cluster for you. When using Helmsman in CI pipeline, these details are required to connect to your cluster everytime the pipeline is executed.

- username   : the username to be used for kubectl credentials.
- password   : an environment variable name (starting with `$`) where your password is stored. Get the password from your k8s admin or consult k8s docs on how to get/set it. 
- clusterURI : the URI for your cluster API or the name of an environment variable (starting with `$`) containing the URI.
- serviceAccount: the name of the service account to use to initiate helm. This should have enough permissions to allow Helm to work and should exist already in the cluster. More details can be found in [helm's RBAC guide](https://github.com/kubernetes/helm/blob/master/docs/rbac.md) 
- storageBackend : by default Helm stores release information in configMaps, using secrets is for storage is recommended for security. Setting this flag to `secret` will deploy/upgrade Tiller with the `--storage=secret`. Other values will be skipped and configMaps will be used.

> If you use `storageBackend` with a Tiller that has been previously deployed with configMaps as storage backend, you need to migrate your release information from the configMap to the new secret on your own. 

Example: 

```toml
[settings]
kubeContext = "minikube" 
# username = "admin"
# password = "$K8S_PASSWORD" 
# clusterURI = "https://192.168.99.100:8443" 
## clusterURI= "$K8S_URI"
# serviceAccount = "my-service-account"
# storageBackend = "secret"
```

```yaml
settings:
  kubeContext = "minikube"
  #username: "admin"
  #password: "$K8S_PASSWORD"
  #clusterURI: "https://192.168.99.100:8443"
  ##clusterURI: "$K8S_URI"
  #serviceAccount: "my-service-account"
  #storageBackend: "secret"
```

## Namespaces

Optional : No.

Synopsis: defines the namespaces to be used/created in your k8s cluster and wether they are protected or not. It also defines if Tiller should be deployed in these namespaces and with what configurations (TLS and service account). You can add as many namespaces as you like.
If a namespaces does not already exist, Helmsman will create it.

Options: 
- protected : defines if a namespace is protected (true or false). Default false.
- installTiller: defines if Tiller should be deployed in this namespace or not. Default is false. Any chart desired to be deployed into a namespace with a Tiller deployed, will be deployed using that Tiller and not the one in kube-system. 
> Tiller will always be deployed into `kube-system`, even if you set installTiller for kube-system to false.

- tillerServiceAccount: defines what service account to use when deploying Tiller. If not set, the `serviceAccount` defined in the `settings` section will be used. If that is also not defined, the namespace `default` service account will be used. If `installTiller` is not defined or set to false, this flag is ignored.
- The following options are `ALL` needed for deploying Tiller with TLS enabled. If they are not all defined, they will be ignored and Tiller will be deployed without TLS. All of these options can be provided as either: a valid local file path, a valid GCS or S3 bucket URI or an environment variable containing a file path or bucket URI.
    - caCert: the CA certificate.
    - tillerCert: the SSL certificate for Tiller.
    - tillerKey: the SSL certificate private key for Tiller.
    - clientCert: the SSL certificate for the Helm client.
    - clientKey: the SSL certificate private key for the Helm client.

> For the defintion of what a protected namespace means, check the [protection guide](how_to/protect_namespaces_and_releases.md)

Example: 

```toml
[namespaces]
[namespaces.staging]
[namespaces.dev]
protected = false
[namespaces.production]
protected = true
installTiller = true
tillerServiceAccount = "tiller-production"
caCert = "secrets/ca.cert.pem"
tillerCert = "secrets/tiller.cert.pem"
tillerKey = "$TILLER_KEY" # where TILLER_KEY=secrets/tiller.key.pem
clientCert = "gs://mybucket/mydir/helm.cert.pem"
clientKey = "s3://mybucket/mydir/helm.key.pem"
```

```yaml
namespaces:
  staging:
  dev:
    protected: false
  production:
    protected: true
    installTiller: true
    tillerServiceAccount: "tiller-production"
    caCert: "secrets/ca.cert.pem"
    tillerCert: "secrets/tiller.cert.pem"
    tillerKey: "$TILLER_KEY" # where TILLER_KEY=secrets/tiller.key.pem
    clientCert: "gs://mybucket/mydir/helm.cert.pem"
    clientKey: "s3://mybucket/mydir/helm.key.pem"
```

## Helm Repos

Optional : No.

Synopsis: defines the Helm repos where your charts can be found. You can add as many repos as you like. Public repos can be added without any additional setup. Private repos require authentication. 

> AS of version v0.2.0, both AWS S3 and Google GCS buckets can be used for private repos (using the [Helm S3](https://github.com/hypnoglow/helm-s3) and [Helm GCS](https://github.com/nouney/helm-gcs) plugins). 

Authenticating to private helm repos:
- **For S3 repos**: you need to have valid AWS access keys in your environment variables. See [here](https://github.com/hypnoglow/helm-s3#note-on-aws-authentication) for more details.
- **For GCS repos**: check [here](https://www.terraform.io/docs/providers/google/index.html#authentication-json-file) for getting the required authentication file. Once you have the file, you have two options, either:
    - set `GOOGLE_APPLICATION_CREDENTIALS` environment variable to contain the absolute path to your Google cloud credentials.json file.
    - Or, set `GCLOUD_CREDENTIALS` environment variable to contain the content of the credentials.json file. 

Options: 
- you can define any key/value pairs where key is the repo name and value is a valid URI for the repo.

Example: 

```toml
[helmRepos]
stable = "https://kubernetes-charts.storage.googleapis.com"
incubator = "http://storage.googleapis.com/kubernetes-charts-incubator"
myS3repo = "s3://my-S3-private-repo/charts"
myGCSrepo = "gs://my-GCS-private-repo/charts"
```

```yaml
helmRepos:
  stable: "https://kubernetes-charts.storage.googleapis.com"
  incubator: "http://storage.googleapis.com/kubernetes-charts-incubator"
  myS3repo: "s3://my-S3-private-repo/charts"
  myGCSrepo: "gs://my-GCS-private-repo/charts"
```

## Apps

Optional : Yes.

Synopsis: defines the releases (instances of Helm charts) you would like to manage in your k8s cluster. 

Releases must have unique names which are defined under `apps`. Example: in `[apps.jenkins]`, the release name will be `jenkins` and it should be unique in your cluster. 

Options: 
- name        : the Helm release name. Releases must have unique names within a cluster.
- description : a release metadata for human readers.
- namespace         : the namespace where the release should be deployed. The namespace should map to one of the ones defined in [namespaces](#namespaces).  
- enabled     : describes the required state of the release (true for enabled, false for disabled). Once a release is deployed, you can change it to false if you want to delete this app release [empty = flase].
- chart       : the chart name. It should contain the repo name as well. Example: repoName/chartName. Changing the chart name means delete and reinstall this release using the new Chart.
- version     : the chart version.
- valuesFile  : a valid path to custom Helm values.yaml file. File extension must be `yaml`. Cannot be used with valuesFiles together. Leaving it empty uses the default chart values.
- valuesFiles : array of valid paths to custom Helm values.yaml file. File extension must be `yaml`. Cannot be used with valuesFile together. Leaving it empty uses the default chart values.
- purge       : defines whether to use the Helm purge flag wgen deleting the release. (true/false)
- test        : defines whether to run the chart tests whenever the release is installed/upgraded/rolledback.
- protected   : defines if the release should be protected against changes. Namespace-level protection has higher priority than this flag. Check the [protection guide](how_to/protect_namespaces_and_releases.md) for more details.
- wait        : defines whether helmsman should block execution until all k8s resources are in a ready state. Default is false.
- priority    : defines the priority of applying operations on this release. Only negative values allowed and the lower the value, the higher the priority. Default priority is 0. Apps with equal priorities will be applied in the order they were added in your state file (DSF).
- [apps.<app_name>.set]  : is used to override certain values from values.yaml with values from environment variables (or ,starting from v1.3.0-rc, directly provided in the Desired State File). This is particularily useful for passing secrets to charts. If the an environment variable with the same name as the provided value exists, the environment variable value will be used, otherwise, the provided value will be used as is.

Example: 

> Whitespace does not matter in TOML files. You could use whatever indentation style you prefer for readability.

```toml
[apps]

    [apps.jenkins]
    name = "jenkins" 
    description = "jenkins"
    namespace = "staging" 
    enabled = true 
    chart = "stable/jenkins" 
    version = "0.9.0"
    valuesFile = "" 
    purge = false 
    test = true 
    protected = false
    wait = true
    priority = -3
    [apps.jenkins.set]
     secret1="$SECRET_ENV_VAR1"
     secret2="SECRET_ENV_VAR2" # works with/without $ at the beginning

```

```yaml
apps:
  jenkins:
    name: "jenkins"
    description: "jenkins"
    namespace: "staging"
    enabled: true
    chart: "stable/jenkins"
    version: "0.9.0"
    valuesFile: ""
    purge: false
    test: true
    protected: false
    wait: true
    priority: -3
    set:
      secret1: "$SECRET_ENV_VAR1"
      secret2: "$SECRET_ENV_VAR2"

```
