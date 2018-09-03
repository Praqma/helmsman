---
version: v1.3.0-rc
---

# Override defined namespaces from command line

If you use different release branches for your releasing/managing your applications in your k8s clusters, then you might want to use the same desired state but with different namespaces on each branch. Instead of duplicating the DSF in multiple branches and adjusting it, you can use the `--ns-override` command line flag when running helmsman.

This flag overrides all namespaces defined in your DSF with the single one you pass from command line.

# Example

dsf.toml
```toml
[metadata]
org = "example.com"
description = "example Desired State File for demo purposes."


[settings]
kubeContext = "minikube"

[namespaces]
  [namespaces.staging]
  protected = false
  [namespaces.production]
  prtoected = true

[helmRepos]
stable = "https://kubernetes-charts.storage.googleapis.com"
incubator = "http://storage.googleapis.com/kubernetes-charts-incubator"


[apps]

    [apps.jenkins]
    name = "jenkins" # should be unique across all apps
    description = "jenkins"
    namespace = "production" # maps to the namespace as defined in environmetns above
    enabled = true # change to false if you want to delete this app release [empty = false]
    chart = "stable/jenkins" # changing the chart name means delete and recreate this chart
    version = "0.14.3" # chart version
    valuesFile = "" # leaving it empty uses the default chart values

    [apps.artifactory]
    name = "artifactory" # should be unique across all apps
    description = "artifactory"
    namespace = "staging" # maps to the namespace as defined in environmetns above
    enabled = true # change to false if you want to delete this app release [empty = false]
    chart = "stable/artifactory" # changing the chart name means delete and recreate this chart
    version = "7.0.6" # chart version
    valuesFile = "" # leaving it empty uses the default chart values
```

dsf.yaml
```yaml
metadata:
  org: "example.com"
  description: "example Desired State File for demo purposes."


settings:
  kubeContext: "minikube"

namespaces:
  staging:
    protected: false
  production:
    protected: true

helmRepos:
  stable: "https://kubernetes-charts.storage.googleapis.com"
  incubator: "http://storage.googleapis.com/kubernetes-charts-incubator"


apps:

  jenkins:
    name: "jenkins" # should be unique across all apps
    description: "jenkins"
    namespace: "production" # maps to the namespace as defined in environments above
    enabled: true # change to false if you want to delete this app release [empty: false]
    chart: "stable/jenkins" # changing the chart name means delete and recreate this chart
    version: "0.14.3" # chart version
    valuesFile: "" # leaving it empty uses the default chart values

  artifactory:
    name: "artifactory" # should be unique across all apps
    description: "artifactory"
    namespace: "staging" # maps to the namespace as defined in environments above
    enabled: true # change to false if you want to delete this app release [empty: false]
    chart: "stable/artifactory" # changing the chart name means delete and recreate this chart
    version: "7.0.6" # chart version
    valuesFile: "" # leaving it empty uses the default chart values
```

In command line, we run :

```
helmsman -f dsf.toml --debug --ns-override testing
```

This will override the `staging` and `production` namespaces defined in `dsf.toml` :

```
2018/03/31 17:38:12 INFO: Plan generated at: Sat Mar 31 2018 17:37:57
DECISION: release [ jenkins ] is not present in the current k8s context. Will install it in namespace [[ testing ]] -- priority: 0
DECISION: release [ artifactory ] is not present in the current k8s context. Will install it in namespace [[ testing ]] -- priority: 0
```
