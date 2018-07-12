---
version: v1.3.0
---

# Multitenant Clusters Guide

This guide helps you use Helmsman to secure your Helm deployment with service accounts and TLS. 

>Checkout Helm's [security guide](https://github.com/kubernetes/helm/blob/master/docs/securing_installation.md)

> These features are available starting from v1.2.0-rc

## Deploying Tiller in multiple namespaces

In a multitenant cluster, it is a good idea to separate the Helm work of different users. You can achieve that by deploying Tiller in multiple namespaces. This is done in the `namespaces` section using the `installTiller` flag:

```toml

[namespaces]
    [namespaces.staging]
    installTiller = true
    [namespaces.production]
    installTiller = true
    [namespaces.developer1]
    installTiller = true
    [namespaces.developer2]
    installTiller = true

```

```yaml

namespaces:
  staging:
    installTiller: true
  production:
    installTiller: true
  developer1:
    installTiller: true
  developer2:
    installTiller: true

```

## Deploying Tiller with a service account 

For K8S clusters with RBAC enabled, you will need to initialize Helm with a service account. Check [Helm's RBAC guide](https://github.com/kubernetes/helm/blob/master/docs/rbac.md).

Helmsman lets you deploy each of the Tillers with a different k8s service account Or with a default service account of your choice. 

```toml

[settings]
# other options
serviceAccount = "default-tiller-sa"

[namespaces]
    [namespaces.staging]
    installTiller = true
    tillerServiceAccount = "custom-sa"

    [namespaces.production]
    installTiller = true
    
    [namespaces.developer1]
    installTiller = true
    tillerServiceAccount = "dev1-sa"

    [namespaces.developer2]
    installTiller = true
    tillerServiceAccount = "dev2-sa"

```

```yaml

settings:
  # other options
  serviceAccount: "default-tiller-sa"

namespaces:
  staging:
    installTiller: true
    tillerServiceAccount: "custom-sa"

  production:
    installTiller: true

  developer1:
    installTiller: true
    tillerServiceAccount: "dev1-sa"

  developer2:
    installTiller: true
    tillerServiceAccount: "dev2-sa"

```
> Currently, Helmsman does not create the service accounts and expects them to be available in the namespace before hand. This should be fixed in upcoming releases and you can track it in [this issue](https://github.com/Praqma/helmsman/issues/48)

> If you don't specify `tillerServiceAccount` option for a namespace, it will try to use the service account you defined in your settings section (`default-tiller-sa` in the example above)

In the example above, namespaces `staging, developer1 & developer2` will have Tiller deployed with different service accounts. 
The `production` namespace ,however, will be deployed using the `default-tiller-sa` service account defined in the `settings` section. If this one is not defined, the production namespace Tiller will be deployed with k8s default service account.

## Deploying Tiller with TLS enabled

In a multitenant setting, it is also recommended to deploy Tiller with TLS enabled. This is also done in the `namespaces` section:

```toml

[namespaces]
    [namespaces.kube-system]
    installTiller = false # has no effect. Tiller is always deployed in kube-system 
    caCert = "secrets/kube-system/ca.cert.pem"
    tillerCert = "secrets/kube-system/tiller.cert.pem"
    tillerKey = "$TILLER_KEY" # where TILLER_KEY=secrets/kube-system/tiller.key.pem
    clientCert = "gs://mybucket/mydir/helm.cert.pem"
    clientKey = "s3://mybucket/mydir/helm.key.pem"

    [namespaces.staging]
    installTiller = true

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
  kube-system:
    installTiller: false # has no effect. Tiller is always deployed in kube-system
    caCert: "secrets/kube-system/ca.cert.pem"
    tillerCert: "secrets/kube-system/tiller.cert.pem"
    tillerKey: "$TILLER_KEY" # where TILLER_KEY=secrets/kube-system/tiller.key.pem
    clientCert: "gs://mybucket/mydir/helm.cert.pem"
    clientKey: "s3://mybucket/mydir/helm.key.pem"

  staging:
    installTiller: true

  production:
    installTiller: true
    tillerServiceAccount: "tiller-production"
    caCert: "secrets/ca.cert.pem"
    tillerCert: "secrets/tiller.cert.pem"
    tillerKey: "$TILLER_KEY" # where TILLER_KEY=secrets/tiller.key.pem
    clientCert: "gs://mybucket/mydir/helm.cert.pem"
    clientKey: "s3://mybucket/mydir/helm.key.pem"

```


