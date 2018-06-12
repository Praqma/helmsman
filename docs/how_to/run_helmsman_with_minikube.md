---
version: v1.2.0-rc
---

You can run Helmsman locally as a binary application with Minikube, you just need to skip all the cluster connection settings in your desired state file. Below is the example.toml desired state file adapted to work with Minikube.


```toml
[metadata]
org = "orgX"
maintainer = "k8s-admin"

[settings]
kubeContext = "minikube" 

[namespaces]
[namespaces.staging]

[helmRepos]
stable = "https://kubernetes-charts.storage.googleapis.com"

[apps]

    [apps.jenkins]
    name = "jenkins" 
    description = "jenkins"
    namespace = "staging" 
    enabled = true 
    chart = "stable/jenkins" 
    version = "0.9.1" 
    valuesFile = "" 
    purge = false 
    test = false 


    [apps.artifactory]
    name = "artifactory" 
    description = "artifactory"
    namespace = "staging" 
    enabled = true 
    chart = "stable/artifactory" 
    version = "6.2.0" 
    valuesFile = "" 
    purge = false 
    test = false 
```

```yaml
metadata:
  org: "orgX"
  maintainer: "k8s-admin"

settings:
  kubeContext: "minikube"

namespaces:
  staging:

helmRepos:
  stable: "https://kubernetes-charts.storage.googleapis.com"

apps:

  jenkins:
    name: "jenkins"
    description: "jenkins"
    namespace: "staging"
    enabled: true
    chart: "stable/jenkins"
    version: "0.9.1"
    valuesFile: ""
    purge: false
    test: false


  artifactory:
    name: "artifactory"
    description: "artifactory"
    namespace: "staging"
    enabled: true
    chart: "stable/artifactory"
    version: "6.2.0"
    valuesFile: ""
    purge: false
    test: false
```