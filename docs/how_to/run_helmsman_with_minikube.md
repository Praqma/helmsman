---
version: v1.1.0
---

You can run Helmsman locally as a binary application with Minikube, you just need to skip all the cluster connection settings in your desired state file. Below is the example.toml desired state file adapted to work with Minikube.


```
[metadata]
org = "orgX"
maintainer = "k8s-admin"

[settings]
kubeContext = "minikube" 

[namespaces]
[namespaces.staging]

[helmRepos]
stable = "https://kubernetes-charts.storage.googleapis.com"
incubator = "http://storage.googleapis.com/kubernetes-charts-incubator"

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