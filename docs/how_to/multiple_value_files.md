---
version: v1.2.0-rc
---

# multiple value files

You can include multiple yaml value files to separate configuration for different environments.

```toml
...
[apps]

    [apps.jenkins]
    name = "jenkins-prod" # should be unique across all apps
    description = "production jenkins"
    namespace = "production"
    enabled = true
    chart = "stable/jenkins"
    version = "0.9.1" # chart version
    valuesFiles = [
        "../my-jenkins-common-values.yaml",
        "../my-jenkins-production-values.yaml"
    ]

    # the jenkins release below is being tested in the staging namespace
    [apps.jenkins-test]
    name = "jenkins-test" # should be unique across all apps
    description = "test release of jenkins, testing xyz feature"
    namespace = "staging" 
    enabled = true 
    chart = "stable/jenkins" 
    version = "0.9.1" # chart version
    valuesFiles = [
        "../my-jenkins-common-values.yaml",
        "../my-jenkins-testing-values.yaml"
    ]

...

```

```yaml
...
apps:

  jenkins:
    name: "jenkins-prod" # should be unique across all apps
    description: "production jenkins"
    namespace: "production"
    enabled: true
    chart: "stable/jenkins"
    version: "0.9.1" # chart version
    valuesFiles:
      - "../my-jenkins-common-values.yaml"
      - "../my-jenkins-production-values.yaml"

  # the jenkins release below is being tested in the staging namespace
  jenkins-test:
    name: "jenkins-test" # should be unique across all apps
    description: "test release of jenkins, testing xyz feature"
    namespace: "staging"
    enabled: true
    chart: "stable/jenkins"
    version: "0.9.1" # chart version
    valuesFiles:
      - "../my-jenkins-common-values.yaml"
      - "../my-jenkins-testing-values.yaml"
...

```