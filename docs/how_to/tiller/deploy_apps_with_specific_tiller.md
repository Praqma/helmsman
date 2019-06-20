---
version: v1.8.0
---

# Deploying apps (releases) with specific Tillers
You can then tell Helmsman to deploy specific releases in a specific namespace:

```toml
#...
[apps]

    [apps.jenkins]
    namespace = "production" # pointing to the namespace defined above
    enabled = true
    chart = "stable/jenkins"
    version = "0.9.1"
    

#...

```

```yaml
# ...
apps:
  jenkins:
    namespace: "production" # pointing to the namespace defined above
    enabled: true
    chart: "stable/jenkins"
    version: "0.9.1"

# ...

```

In the above example, `Jenkins` will be deployed in the production namespace using the Tiller deployed in the production namespace. If the production namespace was not configured to have Tiller deployed there, Jenkins will be deployed using the Tiller in `kube-system`.
