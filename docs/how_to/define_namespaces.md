---
version: v1.2.0-rc
---

# define namespaces

You can define namespaces to be used in your cluster. If they don't exist, Helmsman will create them for you.

```toml
...

[namespaces]
[namespaces.staging]
[namespaces.production]
  protected = true # default is false

...

```

>For details on protecting a namespace, please check the [namespace/release protection guide](protect_namespaces_and_releases.md)

As of `v1.2.0-rc`, you can instruct Helmsman to deploy Tiller into specific namespaces (with or without TLS).

```toml
[namespaces]
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

You can then tell Helmsman to deploy specific releases in a specific namespace:

```toml
...
[apps]

    [apps.jenkins]
    name = "jenkins" 
    description = "jenkins"
    namespace = "production" # pointing to the namespace defined above
    enabled = true 
    chart = "stable/jenkins" 
    version = "0.9.1" 
    valuesFile = "" 
    purge = false 
    test = true  

...

``` 


In the above example, `Jenkins` will be deployed in the production namespace using the Tiller deployed in the production namespace. If the production namespace was not configured to have Tiller deployed there, Jenkins will be deployed using the Tiller in `kube-system`. 

