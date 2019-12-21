---
version: v1.9.0
---

# Limit execution to explicitly defined apps

Starting from v1.9.0, Helmsman allows you to pass the `--target` flag multiple times to specify multiple apps
that limits apps considered by Helmsman during this specific execution. 
Thanks to this one can deploy specific applications among all defined for an environment.

## Example

Having environment defined with such apps:

* example.yaml:
```yaml
# ...
apps:
    jenkins:
      namespace: "staging" # maps to the namespace as defined in namespaces above
      enabled: true # change to false if you want to delete this app release empty: false:
      chart: "stable/jenkins" # changing the chart name means delete and recreate this chart
      version: "0.14.3" # chart version

    artifactory:
      namespace: "production" # maps to the namespace as defined in namespaces above
      enabled: true # change to false if you want to delete this app release empty: false:
      chart: "stable/artifactory" # changing the chart name means delete and recreate this chart
      version: "7.0.6" # chart version
# ...
```

running Helmsman with `-f example.yaml` would result in checking state and invoking deployment for both jenkins and artifactory application.

With `--target` flag in command like

```shell
$ helmsman -f example.yaml --target artifactory ...
```

one can execute Helmsman's environment defined with example.yaml limited to only one `artifactory` app. Others are ignored until the flag is defined.

Multiple applications can be set with `--target`, like

```shell
$ helmsman -f example.yaml --target artifactory --target jenkins ...
```
