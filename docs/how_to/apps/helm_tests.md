---
version: v1.3.0-rc
---

# Test charts

Helm allows running [chart tests](https://github.com/helm/helm/blob/master/docs/chart_tests.md). 

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
# ...
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

#...

```
