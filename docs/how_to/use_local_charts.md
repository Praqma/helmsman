---
version: v1.3.0-rc
---

# use local helm charts

You can use your locally developed charts. 

## Served by Helm

You can serve them on localhost using helm's `serve` option.

```toml
...

[helmRepos]
stable = "https://kubernetes-charts.storage.googleapis.com"
incubator = "http://storage.googleapis.com/kubernetes-charts-incubator"
local = http://127.0.0.1:8879

...

```

```yaml
...

helmRepos:
  stable: "https://kubernetes-charts.storage.googleapis.com"
  incubator: "http://storage.googleapis.com/kubernetes-charts-incubator"
  local: http://127.0.0.1:8879

...

```

## From file system

If you use a file path (relative to the DSF, or absolute) for the ```chart``` attribute
helmsman will try to resolve that chart from the local file system. The chart on the 
local file system must have a version matching the version specified in the DSF.

