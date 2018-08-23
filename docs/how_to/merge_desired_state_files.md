---
version: v1.5.0-rc
---

# supply multiple desired state files

Starting from v1.5.0-rc, Helmsman allows you to pass the `-f` flag multiple times to specify multiple desired state files
that should be merged. This allows us to do things like specify our non-environment-specific config in a `common.toml` file
and environment specific info in a `nonprod.toml` or `prod.toml` file. This process uses [this library](https://github.com/imdario/mergo)
to do the merging, and is subject to the limitations described there.

For example:

* common.toml:
```toml
[metadata]
org = "Organization Name"
maintainer = "project-owners@example.com"
description = "Project charts"

[settings]
serviceAccount = "tiller"
storageBackend = "secret"
...
```

* nonprod.toml:
```toml
[settings]
kubeContext = "cluster-nonprod"

[apps]
  [apps.external-dns]
  valuesFiles = ["./external-dns/values.yaml", "./external-dns/nonprod.yaml"]

  [apps.cert-issuer]
  valuesFile = "./cert-issuer/nonprod.yaml"
...
```

One can then run the following to use the merged config of the above files, with later files override values of earlier ones:
```bash
$ helmsman -f common.toml -f nonprod.toml ...
```
