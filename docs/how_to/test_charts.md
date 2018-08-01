---
version: v1.3.0-rc
---

# test charts

You can specify that you would like a chart to be tested whenever it is installed for the first time using the `test` key as follows:

```toml
...
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
    test = true  # setting this to true, means you want the charts tests to be run on this release when it is installed. 

...

```

```yaml
...
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
    test: true  # setting this to true, means you want the charts tests to be run on this release when it is installed.

...

```