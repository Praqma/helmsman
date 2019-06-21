---
version: v1.8.0
---

# Deploying multiple Tillers

You can deploy multiple Tillers in the cluster (max. one per namespace). In each namespace definition you can configure how Tiller is installed. The following options are available:
- with/without RBAC
- with/without TLS
- with cluster-admin clusterrole or with a namespace-limited role or with an pre-configured role.

> If you use GCS, S3, or Azure blob storage for your certificates, you will need to provide means to authenticate to the respective cloud provider in the environment. See [authenticating to cloud storage providers](../misc/auth_to_storage_providers.md) for details.


> More details about using Helmsman in a multitenant cluster can be found [here](../misc/multitenant_clusters_guide.md)

You can also use pre-configured Tillers in specific namespaces. In the example below, the desired state is: to deploy Tiller in the `production` namespace with TLS and RBAC, and to use a pre-configured Tiller in the `dev` namespace. The `staging` namespace does not have any Tiller to be deployed or used. Tiller is not deployed in `kube-system`.


```toml
[namespaces]
   # to prevent deploying Tiller into kube-system, use the two lines below
  [namespaces.kube-system]
    installTiller = false # this line can be omitted since installTiller defaults to false
  [namespaces.staging]
  [namespaces.dev]
    useTiller = true # use a Tiller which has been deployed in dev namespace
  [namespaces.production]
    installTiller = true
    tillerServiceAccount = "tiller-production"
    tillerRole = "cluster-admin"
    caCert = "secrets/ca.cert.pem"
    tillerCert = "az://myblobcontainer/tiller.cert.pem"
    tillerKey = "$TILLER_KEY" # where TILLER_KEY=secrets/tiller.key.pem
    clientCert = "gs://mybucket/mydir/helm.cert.pem"
    clientKey = "s3://mybucket/mydir/helm.key.pem"
```

```yaml
namespaces:
  # to prevent deploying Tiller into kube-system, use the two lines below
  kube-system:
   installTiller: false # this line can be omitted since installTiller defaults to false
  staging: # no Tiller deployed or used here
  dev:
    useTiller: true # use a Tiller which has been deployed in dev namespace
  production:
    installTiller: true
    tillerServiceAccount: "tiller-production"
    tillerRole: "cluster-admin"
    caCert: "secrets/ca.cert.pem"
    tillerCert: "az://myblobcontainer/tiller.cert.pem"
    tillerKey: "$TILLER_KEY" # where TILLER_KEY=secrets/tiller.key.pem
    clientCert: "gs://mybucket/mydir/helm.cert.pem"
    clientKey: "s3://mybucket/mydir/helm.key.pem"
```
