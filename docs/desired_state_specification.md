---
version: v1.12.0
---

# Helmsman desired state specification

This document describes the specification for how to write your Helm charts' desired state file. This can be either a [Toml](https://github.com/toml-lang/toml) or [Yaml](http://yaml.org/) formatted file. The desired state file consists of:

- [Metadata](#metadata) [Optional] -- metadata for any human reader of the desired state file.
- [Certificates](#certificates) [Optional] -- only needed when you want Helmsman to connect kubectl to your cluster for you.
- [Context](#context) [optional] -- define the context in which a DSF is used.
- [Settings](#settings) [Optional] -- data about your k8s cluster and how to deploy Helm on it if needed.
- [Namespaces](#namespaces) -- defines the namespaces where you want your Helm charts to be deployed.
- [Helm Repos](#helm-repos) [Optional] -- defines the repos where you want to get Helm charts from.
- [Apps](#apps) -- defines the applications/charts you want to manage in your cluster.


> You can use environment variables in the desired state files. The environment variable name should start with "$", or encapsulated in "${", "}". "$" characters can be escaped like "$$".

> Starting from v1.9.0, you can also use environment variables in your helm values/secrets files.

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

Synopsis: defines where to find the certificates needed for connecting kubectl to a k8s cluster. If connection settings (username/password/clusterAPI) are provided in the Settings section below, then you need **AT LEAST** to provide caCrt and caKey. You can optionally provide a client certificate (caClient) depending on your cluster connection setup.

Options:
- **caCrt** : a valid S3/GCS/Azure bucket or local relative file path to a certificate file.
- **caKey** : a valid S3/GCS/Azure bucket or local relative file path to a client key file.
- **caClient**: a valid S3/GCS/Azure bucket or local relative file path to a client certificate file.


> bucket format is: [s3 or gs or az]://bucket-name/dir1/dir2/.../file.extension

Example:

```toml
[certificates]
caCrt = "s3://myS3bucket/mydir/ca.crt"
caKey = "gs://myGCSbucket/ca.key"
#caKey = "az://myAzureContainer/ca.key
caClient ="../path/to/my/local/client-certificate.crt"
#caClient = "$CA_CLIENT"
```

```yaml
certificates:
  caCrt: "s3://myS3bucket/mydir/ca.crt"
  caKey: "gs://myGCSbucket/ca.key"
  #caKey:  "az://myAzureContainer/ca.key
  caClient: "../path/to/my/local/client-certificate.crt"
  #caClient: "$CA_CLIENT"
```
## Context

Optional : Yes.

Synopsis: defines the context in which a DSF is used. This context is used as the ID of that specific DSF and must be unique across the used DSFs. If not defined, `default` is used. Check [here](how_to/misc/merge_desired_state_files.md) for more details on the limitations.

```yaml
context: prod-apps
...
```  

## Settings

Optional : Yes.

Synopsis: provides settings for connecting to your k8s cluster and configuring Helm's Tiller in the cluster.

> If you don't provide the `settings` stanza, helmsman would use your current kube context and will deploy Tiller(s) without RBAC service accounts.

Options:
- **kubeContext** : the kube context you want Helmsman to use or create. Helmsman will try connect to this context first, if it does not exist, it will try to create it (i.e. connect to a k8s cluster) using the options below.

The following options can be skipped if your kubectl context is already created and you don't want Helmsman to connect kubectl to your cluster for you.

- **username**   : the username to be used for kubectl credentials.
- **password**   : an environment variable name (starting with `$`) where your password is stored. Get the password from your k8s admin or consult k8s docs on how to get/set it.
- **clusterURI** : the URI for your cluster API or the name of an environment variable (starting with `$`) containing the URI.
- **bearerToken**: whether you want helmsman to connect to the cluster using a bearer token. Default is `false`
- **bearerTokenPath**: optional. If bearer token is used, you can specify a custom location for the token file.
- **serviceAccount**: the name of the service account to use to deploy Helm Tiller. This should have enough permissions to allow Helm to work. If the service account does not exist, it will be created. More details can be found in [helm's RBAC guide](https://github.com/kubernetes/helm/blob/master/docs/rbac.md)
- **storageBackend** : by default Helm stores release information in configMaps, using secrets is for storage is recommended for security. Setting this flag to `secret` will deploy/upgrade Tiller with the `--storage=secret`. Other values will be skipped and configMaps will be used.
- **slackWebhook** : a [Slack](http://slack.com) Webhook URL to receive Helmsman notifications. This can be passed directly or in an environment variable.
- **reverseDelete** : if set to `true` it will reverse the priority order whilst deleting.
- **tillerless** : setting it to `true` will use [helm-tiller](https://rimusz.net/tillerless-helm) plugin instead of installing Tillers in namespaces. It disables many of the parameters for sections below.
- **eyamlEnabled** : if set to `true' it will use [hiera-eyaml](https://github.com/voxpupuli/hiera-eyaml) to decrypt secret files instead of using default helm-secrets based on sops
- **eyamlPrivateKeyPath** : if set with path to the eyaml private key file, it will use it instead of looking for default one in ./keys directory relative to where Helmsman were run. It needs to be defined in conjunction with eyamlPublicKeyPath.
- **eyamlPublicKeyPath** : if set with path to the eyaml public key file, it will use it instead of looking for default one in ./keys directory relative to where Helmsman were run. It needs to be defined in conjunction with eyamlPrivateKeyPath.

> If you use `storageBackend` with a Tiller that has been previously deployed with configMaps as storage backend, you need to migrate your release information from the configMap to the new secret on your own. Helm does not support this yet.

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
# slackWebhook = $MY_SLACK_WEBHOOK
# reverseDelete = false
# tilerless = true
# eyamlEnabled = true
# eyamlPrivateKeyPath = "../keys/custom-key.pem"
# eyamlPublicKeyPath = "../keys/custom-key.pub"
```

```yaml
settings:
  kubeContext: "minikube"
  #username: "admin"
  #password: "$K8S_PASSWORD"
  #clusterURI: "https://192.168.99.100:8443"
  ##clusterURI: "$K8S_URI"
  #serviceAccount: "my-service-account"
  #storageBackend: "secret"
  #slackWebhook: "$MY_SLACK_WEBHOOK"
  #reverseDelete: false
  #tilerless: true
  # eyamlEnabled: true
  # eyamlPrivateKeyPath: ../keys/custom-key.pem
  # eyamlPublicKeyPath: ../keys/custom-key.pub
```

## Namespaces

Optional : No.

Synopsis: defines the namespaces to be used/created in your k8s cluster and whether they are protected or not. It also defines if Tiller should be deployed in these namespaces and with what configurations (TLS and service account). You can add as many namespaces as you like.
If a namespace does not already exist, Helmsman will create it.

> All Tillers-related params here will be ignored when `tilerless` is set in `settings` section.

Options:
- **protected** : defines if a namespace is protected (true or false). Default false.
> For the definition of what a protected namespace means, check the [protection guide](how_to/misc/protect_namespaces_and_releases.md)
- **installTiller**: defines if Tiller should be deployed in this namespace or not. Default is false. Any chart desired to be deployed into a namespace with a Tiller deployed, will be deployed using that Tiller and not the one in kube-system unless you use the `TillerNamespace` option (see the [Apps](#apps) section below) to use another Tiller.
> By default Tiller will be deployed into `kube-system` even if you don't define kube-system in the namespaces section. To prevent deploying Tiller into `kube-system, add kube-system in your namespaces section and set its installTiller to false.
- **tillerMaxHistory**: specifies an int value of the maximum number of revisions saved per release by Tiller.
> In order to set the kube-system's Tiller's (a default one, main Tiller) max history, define namespace kube-system and set tillerMaxHistory along with installTiller: true
- **tillerRole**: specifies the role to use.  If 'cluster-admin' a clusterrolebinding will be used else a role with a single namespace scope will be created and bound with a rolebinding.
- **tillerRoleTemplateFile**: relative path to file templating custom Tiller role. If `installTiller` is true and `tillerRole` is not `cluster-admin`, then helmsman will create namespace specific Tiller role based on the template file passed with this parameter. When `tillerRole` is empty string, role name defaults to `helmsman-tiller`.

  If `tillerRoleTemplateFile` is set, it will always try to create namespace-scoped Tiller with conditions like below:
    - if `tillerRole` was not set, default name of `helmsman-tiller` will be used to createa new Role
    - if `tillerServiceAccount` was not set, it will create Service Account with default name of `helmsman`

- **useTiller**: defines that you would like to use an existing Tiller from that namespace. Can't be set together with `installTiller`
- **labels** : defines labels to be added to the namespace, doesn't remove existing labels but updates them if the label key exists with any other different value. You can define any key/value pairs. Default is empty.
- **annotations** : defines annotations to be added to the namespace. It behaves the same way as the labels option.
- **limits** : defines a [LimitRange](https://kubernetes.io/docs/tasks/administer-cluster/manage-resources/memory-default-namespace/) to be configured on the namespace

- **tillerServiceAccount**: defines what service account to use when deploying Tiller. If this is not set, the following options are considered:

  1. If the `serviceAccount` defined in the `settings` section exists in the namespace you want to deploy Tiller in, it will be used. 
  2. If you defined `serviceAccount` in the `settings` section and it does not exist in the namespace you want to deploy Tiller in, Helmsman creates the service account in that namespace and binds it to a (cluster)role. If the namespace is kube-system and `tillerRole` is unset or is set to cluster-admin, the service account is bound to `cluster-admin` clusterrole. Otherwise, if you specified a `tillerRole`, a new role with that name is created and bound to the service account with rolebinding. If `tillerRole` is unset (for namespaces other than kube-system), the role is called `helmsman-tiller` and is created in the specified namespace to only gives access to that namespace. The custom role is created from a [yaml template](../data/role.yaml) if no `tillerRoleTemplateFile` was set, or it will use the `tillerRoleTemplateFile` (of the same structure as [yaml template](../data/role.yaml)) to create Role from.

  > If `installTiller` is not defined or set to false, this flag is ignored.

- The following options are `ALL` needed for deploying Tiller with TLS enabled. If they are not all defined, they will be ignored and Tiller will be deployed without TLS. All of these options can be provided as either: a valid local file path, a valid GCS or S3 or Azure bucket URI or an environment variable containing a file path or bucket URI.
    - **caCert**: the CA certificate.
    - **tillerCert**: the SSL certificate for Tiller.
    - **tillerKey**: the SSL certificate private key for Tiller.
    - **clientCert**: the SSL certificate for the Helm client.
    - **clientKey**: the SSL certificate private key for the Helm client.

Example:

```toml
[namespaces]
# to prevent deploying Tiller into kube-system, use the two lines below
# [namespaces.kube-system]
# installTiller = false # this line can be omitted since installTiller defaults to false
[namespaces.staging]
[namespaces.dev]
useTiller = true # use a Tiller which has been deployed in dev namespace
protected = false
[namespaces.production]
protected = true
installTiller = true
tillerMaxHistory = 10
tillerServiceAccount = "tiller-production"
tillerRoleTemplateFile = "../roles/helmsman-tiller.yaml"
caCert = "secrets/ca.cert.pem"
tillerCert = "secrets/tiller.cert.pem"
tillerKey = "$TILLER_KEY" # where TILLER_KEY=secrets/tiller.key.pem
clientCert = "gs://mybucket/mydir/helm.cert.pem"
clientKey = "s3://mybucket/mydir/helm.key.pem"
[namespaces.production.labels]
env = "prod"
[namespaces.production.annotations]
iam.amazonaws.com/role = "dynamodb-reader"
[[namespaces.production.limits]]
type = "Container"
[namespaces.production.limits.default]
cpu = "300m"
memory = "200Mi"
[namespaces.production.limits.defaultRequest]
cpu = "200m"
memory = "100Mi"
[[namespaces.production.limits]]
type = "Pod"
[namespaces.production.limits.max]
memory = "300Mi"
```

```yaml
namespaces:
  # to prevent deploying Tiller into kube-system, use the two lines below
  # kube-system:
  #  installTiller: false # this line can be omitted since installTiller defaults to false
  staging:
  dev:
    protected: false
    useTiller: true # use a Tiller which has been deployed in dev namespace
  production:
    protected: true
    installTiller: true
    tillerMaxHistory: 10
    tillerServiceAccount: "tiller-production"
    tillerRoleTemplateFile: "../roles/helmsman-tiller.yaml"
    caCert: "secrets/ca.cert.pem"
    tillerCert: "secrets/tiller.cert.pem"
    tillerKey: "$TILLER_KEY" # where TILLER_KEY=secrets/tiller.key.pem
    clientCert: "gs://mybucket/mydir/helm.cert.pem"
    clientKey: "s3://mybucket/mydir/helm.key.pem"
    limits:
      - type: Container
        default:
          cpu: "300m"
          memory: "200Mi"
        defaultRequest:
          cpu: "200m"
          memory: "100Mi"
      - type: Pod
        max:
          memory: "300Mi"
    labels:
      env: "prod"
    annotations:
      iam.amazonaws.com/role: "dynamodb-reader"
```

## Helm Repos

Optional : Yes.

Synopsis: defines the Helm repos where your charts can be found. You can add as many repos as you like. Public repos can be added without any additional setup. Private repos require authentication.

> As of version v0.2.0, both AWS S3 and Google GCS buckets can be used for private repos (using the [Helm S3](https://github.com/hypnoglow/helm-s3) and [Helm GCS](https://github.com/nouney/helm-gcs) plugins).

> As of version v1.8.0, you can use private repos with basic auth and you can use pre-configured helm repos.

Authenticating to private helm repos:
- **For S3 repos**: you need to have valid AWS access keys in your environment variables. See [here](https://github.com/hypnoglow/helm-s3#note-on-aws-authentication) for more details.
- **For GCS repos**: check [here](https://www.terraform.io/docs/providers/google/index.html#authentication-json-file) for getting the required authentication file. Once you have the file, you have two options, either:
    - set `GOOGLE_APPLICATION_CREDENTIALS` environment variable to contain the absolute path to your Google cloud credentials.json file.
    - Or, set `GCLOUD_CREDENTIALS` environment variable to contain the content of the credentials.json file.

Options:
- you can define any key/value pair where the key is the repo name and value is a valid URI for the repo. Basic auth info can be added in the repo URL as in the example below.

Example:

```toml
[helmRepos]
stable = "https://kubernetes-charts.storage.googleapis.com"
incubator = "http://storage.googleapis.com/kubernetes-charts-incubator"
myS3repo = "s3://my-S3-private-repo/charts"
myGCSrepo = "gs://my-GCS-private-repo/charts"
myPrivateRepo = "https://user:$TOP_SECRET_PASSWORD@mycustomprivaterepo.org"
```

```yaml
helmRepos:
  stable: "https://kubernetes-charts.storage.googleapis.com"
  incubator: "http://storage.googleapis.com/kubernetes-charts-incubator"
  myS3repo: "s3://my-S3-private-repo/charts"
  myGCSrepo: "gs://my-GCS-private-repo/charts"
  myPrivateRepo: "https://user:$TOP_SECRET_PASSWORD@mycustomprivaterepo.org"
```

## Preconfigured Helm Repos

Optional : Yes.

Synopsis: defines the list of helm repositories that the helmsman will consider already preconfigured and thus will not try to overwrite it's configuration.

The primary use-case is if you have some helm repositories that require HTTP basic authentication and you don't want to store the password in the desired state file or as an environment variable. In this case you can execute the following sequence to have those repositories configured:

Set up the helmsman configuration:

```toml
preconfiguredHelmRepos = [ "myrepo1", "myrepo2" ]
```

```yaml
preconfiguredHelmRepos:
- myrepo1
- myrepo2
```

> In this case you will manually need to execute `helm repo add myrepo1 <URL> --username= --password=`

## AppsTemplates

Optional : Yes.

Synopsis: allows for YAML (TOML has no variable reference support) object creation, that is ignored by state file importer, but can be used as a reference with YAML anchors to not repeat yourself. Read [this](https://blog.daemonl.com/2016/02/yaml.html) example about YAML anchors.

Examples:

```yaml
appsTemplates:

  default: &template
    valuesFile: ""
    purge: false
    test: true
    protected: false
    wait: true
    enabled: true

  custom: &template_custom
    valuesFile: ""
    purge: true
    test: true
    protected: false
    wait: false
    enabled: true

apps:
  jenkins:
    <<: *template
    name: "jenkins-stage"
    namespace: "staging"
    chart: "stable/jenkins"
    version: "0.9.2"
    priority: -3

  jenkins2:
    <<: *template_custom
    name: "jenkins-prod"
    namespace: "production"
    chart: "stable/jenkins"
    version: "0.9.0"
    priority: -2

```

## Apps

Optional : Yes.

Synopsis: defines the releases (instances of Helm charts) you would like to manage in your k8s cluster.

Releases must have unique names which are defined under `apps`. Example: in `[apps.jenkins]`, the release name will be `jenkins` and it should be unique within the Tiller which manages it .

Options:

**Required**
- **namespace**         : the namespace where the release should be deployed. The namespace should map to one of the ones defined in [namespaces](#namespaces).
- **enabled**     : describes the required state of the release (true for enabled, false for disabled). Once a release is deployed, you can change it to false if you want to delete this release [default is false].
- **chart**       : the chart name. It should contain the repo name as well. Example: repoName/chartName. Changing the chart name means delete and reinstall this release using the new Chart.
- **version**     : the chart version.

**Optional**
- **group**       : group name this apps belongs to. It has no effect until Helmsman's flag `-group` is passed. Check this [doc](how_to/misc/limit-deployment-to-specific-group-of-apps.md) for more details.
- **tillerNamespace** : which Tiller to use for deploying this release. This is available starting from v1.4.0-rc The decision on which Tiller to use for deploying a release follows the following criteria:
   1. If `tillerNamespace`is explicitly defined, it is used.
   2. If `tillerNamespace`is not defined and the namespace in which the release will be deployed has a Tiller installed by Helmsman (i.e. has `installTiller set to true` in the [Namespaces](#namespaces) section), Tiller in that namespace is used.
   3. If none of the above, the shared Tiller in `kube-system` is used.
> This parameter is ignored when `tilerless` is set in `settings` section and `namespace` of this app is taken for tilerless plugin.

- **name**        : the Helm release name. Releases must have unique names within a Helm Tiller. If not set, the release name will be taken from the app identifier in your desired state file. e.g, for ` apps.jenkins ` the release name will be `jenkins`.
- **description** : a release metadata for human readers.
- **valuesFile**  : a valid path to custom Helm values.yaml file. File extension must be `yaml`. Cannot be used with valuesFiles together. Leaving it empty uses the default chart values.
- **valuesFiles** : array of valid paths to custom Helm values.yaml file. File extension must be `yaml`. Cannot be used with valuesFile together. Leaving it empty uses the default chart values.
> The values file(s) path is resolved when the DSF yaml/toml file is loaded, relative to the path that the dsf was loaded from.
- **secretsFile**  : a valid path to custom Helm secrets.yaml file. File extension must be `yaml`. Cannot be used with secretsFiles together. Leaving it empty uses the default chart secrets.
- **secretsFiles** : array of valid paths to custom Helm secrets.yaml file. File extension must be `yaml`. Cannot be used with secretsFile together. Leaving it empty uses the default chart secrets.
> The secrets file(s) path is resolved when the DSF yaml/toml file is loaded, relative to the path that the dsf was loaded from.
> To use the secrets files you must have the helm-secrets plugin
- **purge**       : defines whether to use the Helm purge flag when deleting the release. Default is false.
- **test**        : defines whether to run the chart tests whenever the release is installed. Default is false.
- **protected**   : defines if the release should be protected against changes. Namespace-level protection has higher priority than this flag. Check the [protection guide](how_to/misc/protect_namespaces_and_releases.md) for more details. Default is false.
- **wait**        : defines whether Helmsman should block execution until all k8s resources are in a ready state. Default is false.
- **timeout**     : helm timeout in seconds. Default 300 seconds.
- **noHooks**     : helm noHooks option. If true, it will disable pre/post upgrade hooks. Default is false.
- **priority**    : defines the priority of applying operations on this release. Only negative values allowed and the lower the value, the higher the priority. Default priority is 0. Apps with equal priorities will be applied in the order they were added in your state file (DSF).
- **set**  : is used to override certain values from values.yaml with values from environment variables (or ,starting from v1.3.0-rc, directly provided in the Desired State File). This is particularly useful for passing secrets to charts. If the an environment variable with the same name as the provided value exists, the environment variable value will be used, otherwise, the provided value will be used as is. The TOML stanza for this is `[apps.<app_name>.set]`
- **setString**   : is used to override String values from values.yaml or chart's defaults. This uses the `--set-string` flag in helm which is available only in helm >v2.9.0. This option is useful for image tags and the like. The TOML stanza for this is `[apps.<app_name>.setString]`
- **helmFlags**   : array of `helm` flags, is used to pass flags to helm install/upgrade commands

Example:

> Whitespace does not matter in TOML files. You could use whatever indentation style you prefer for readability.

```toml
[apps]

  [apps.jenkins]
    name = "jenkins"
    description = "jenkins"
    namespace = "staging"
    enabled = true
    group = "critical"
    chart = "stable/jenkins"
    version = "0.9.0"
    valuesFile = ""
    purge = false
    test = true
    protected = false
    wait = true
    priority = -3
    helmFlags = [
      "--recreate-pods",
    ]
  [apps.jenkins.set]
    secret1="$SECRET_ENV_VAR1"
    secret2="SECRET_ENV_VAR2" # works with/without $ at the beginning
  [apps.jenkins.setString]
    longInt = "1234567890"
    "image.tag" = "1.0.0"
```

```yaml
apps:
  jenkins:
    name: "jenkins"
    description: "jenkins"
    namespace: "staging"
    enabled: true
    group: "critical"
    chart: "stable/jenkins"
    version: "0.9.0"
    valuesFile: ""
    purge: false
    test: true
    protected: false
    wait: true
    priority: -3
    helmFlags: [
      "--recreate-pods",
    ]
    set:
      secret1: "$SECRET_ENV_VAR1"
      secret2: "$SECRET_ENV_VAR2"
    setString:
      longInt: "1234567890"
      image:
        tag: "1.0.0"
```
