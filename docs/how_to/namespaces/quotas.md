---
version: ???
---

# Define resource quotas for namespaces

You can define namespaces to be used in your cluster. If they don't exist, Helmsman will create them for you. You can also define how much resource limits to set for each namespace.

You can read more about the `Quotas` specification [here](https://kubernetes.io/docs/tasks/administer-cluster/manage-resources/quota-memory-cpu-namespace/#create-a-resourcequota).

```toml
#...
[namespaces]
  [namespaces.helmsman1]
    [namespaces.helmsman1.quotas]
    pods = "25"
      [namespaces.helmsman1.quotas.requests]
      memory = "5Gi"
      cpu = "10"
      [namespaces.helmsman1.quotas.limits]
      memory = "5Gi"
      cpu = "10"
      [[namespaces.helmsman1.quotas.customQuotas]]
      name = "requests.nvidia.com/gpu"
      value = 0.0
#...
```

```yaml
namespaces:
  helmsman1:
    quotas:
      pods: "25"
      requests:
        memory: "5Gi"
        cpu: "10"
      limits:
        memory: "5Gi"
        cpu: "10"
      customQuotas:
      - name: requests.nvidia.com/gpu
        value: 0
```

The example above will create two namespaces - staging and production - with resource limits defined for the staging namespace.