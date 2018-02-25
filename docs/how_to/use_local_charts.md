---
version: v1.0.0
---

# use local helm charts

You can use your locally developed charts. But first, you have to serve them on localhost using helm's `serve` option.

```
...

[helmRepos]
stable = "https://kubernetes-charts.storage.googleapis.com"
incubator = "http://storage.googleapis.com/kubernetes-charts-incubator"
local = http://127.0.0.1:8879

...
``` 