---
version: v1.8.0
---

# Prevent Tiller Deployment in kube-system

By default Tiller will be deployed into `kube-system` even if you don't define kube-system in the namespaces section. To prevent this, simply add `kube-system` into your namespaces section. Since `installTiller` for namespaces is by default false, Helmsman will not deploy Tiller in `kube-system`.

```toml
[namespaces]
[namespaces.kube-system]
# installTiller = false  # this line is not needed since the default is false, but can be added for human readability.
```

```yaml
namespaces:
  kube-system:
    #installTiller: false # this line is not needed since the default is false, but can be added for human readability.
```
