---
version: v3.7.1
---

# Using OCI registries for helm charts

Helmsman allows you to use charts stored in OCI registries.

You need to export the following env variables:

- `HELM_EXPERIMENTAL_OCI=1`

if the registry requires authentication, you must login before running Helmsman

```sh
helm registry login -u myuser my-registry.local
```

```toml
[apps]
  [apps.my-app]
    chart = "oci://my-registry.local/my-chart"
    version = "1.0.0"
```

```yaml
#...
apps:
  my-app:
    chart: oci://my-registry.local/my-chart
    version: 1.0.0
```

For more information, read the [helm registries documentation](https://helm.sh/docs/topics/registries/).
