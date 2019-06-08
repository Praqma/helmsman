---
version: v1.10.0
---

# Create helmsman per namespace with your custom Role

You can deploy namespace-specific helmsman Tiller with Service Account having custom Role.

By default, when defining Namespaces in Desired State Specification file, when `installTiller` is enabled for specific namespace,
it creates the Role to bind the Tiller to with default [yaml template](../../../data/role.yaml). 

If there's a need for custom Role (let's say each namespace has its different and specific requirements to permissions),
you can define `tillerRoleConfigFile`, which is a relative path pointing at a template of a Role (same format as a [yaml template](../../../data/role.yaml)),
so when Helmsman creates Tiller in the namespace with this key, custom Role will be created for Tiller.

```toml
[namespaces]
[namespaces.dev]
useTiller = true
[namespaces.production]
installTiller = true
tillerServiceAccount = "tiller-production"
tillerRoleConfigFile = "../roles/helmsman-tiller.yaml"
```

```yaml
namespaces:
  dev:
    useTiller: true
  production:
    installTiller: true
    tillerServiceAccount: "tiller-production"
    tillerRoleConfigFile: "../roles/helmsman-tiller.yaml"
```

The example above will create two namespaces: dev and production, where dev namespace will have its Tiller with default Role,
while production namespace will be managed by its specific Tiller having custom role based on the `"../roles/helmsman-tiller.yaml"` template created by you.
