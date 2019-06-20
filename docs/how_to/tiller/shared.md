---
version: v1.8.0
---

# Deploying a shared Tiller (available from v1.2.0)

You can instruct Helmsman to deploy Tiller into specific namespaces (with or without TLS).

> By default Tiller will be deployed into `kube-system` even if you don't define kube-system in the namespaces section. To prevent deploying Tiller into `kube-system, see [preventing Tiller deployment in kube-system](prevent_tiller_in_kube_system.md)

## Without TLS 

```toml
[namespaces]
[namespaces.production]
  installTiller = true
```

```yaml
namespaces:
  production:
    installTiller: true
```

## With RBAC service account

You specify an existing service account to be used for deploying Tiller. If that service account does not exist, Helmsman will attempt to create it. If `tillerRole` (e.g. cluster-admin) is specified, it will be bound to the newly created service account.

By default, Tiller deployed in kube-system will be given cluster-admin clusterrole. Tiller in other namespaces will be given a custom role that gives it access to that namespace only. The custom role is created using [this template](../../../data/role.yaml). 

```toml
[namespaces]
[namespaces.production]
  installTiller = true
  tillerServiceAccount = "tiller-production"
  tillerRole = "cluster-admin"
[namespaces.staging]
  installTiller = true
  tillerServiceAccount = "tiller-stagin"  
```

```yaml
namespaces:
  production:
    installTiller: true
    tillerServiceAccount: "tiller-production"
    tillerRole: "cluster-admin"
  staging:
    installTiller: true
    tillerServiceAccount: "tiller-staging"  
```

The above example will create two service accounts; `tiller-production` and `tiller-staging`. Service account `tiller-production` will be bound to a cluster admin clusterrole while `tiller-staging` will be bound to a newly created role with access to the staging namespace only. 

## With RBAC and TLS

You have to provide the TLS certificates as below. Certificates can be either located locally or in Google GCS, AWS S3 or Azure blob storage.

> If you use GCS, S3, or Azure blob storage for your certificates, you will need to provide means to authenticate to the respective cloud provider in the environment. See [authenticating to cloud storage providers](../misc/auth_to_storage_providers.md) for details.

```toml
[namespaces]
[namespaces.production]
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
  production:
    installTiller: true
    tillerServiceAccount: "tiller-production"
    caCert: "secrets/ca.cert.pem"
    tillerCert: "secrets/tiller.cert.pem"
    tillerKey: "$TILLER_KEY" # where TILLER_KEY=secrets/tiller.key.pem
    clientCert: "gs://mybucket/mydir/helm.cert.pem"
    clientKey: "s3://mybucket/mydir/helm.key.pem"
```
