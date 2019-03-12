---
version: v1.8.0
---

## Using your existing Tillers (available from v1.6.0)

If you would like to use custom configuration when deploying your Tiller, you can do that before using Helmsman and then use the `useTiller` option in your namespace definition.

This will allow Helmsman to use your existing Tiller as it is. Note that you can't set both `useTiller` and `installTiller` to true at the same time.

```toml
[namespaces]
[namespaces.production]
  useTiller = true
```

```yaml
namespaces:
  production:
    useTiller: true
```