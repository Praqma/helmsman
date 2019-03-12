---
version: v1.8.0
---

# Default helm repos

By default, helm comes with two default repos; `stable` and `incubator`. These two DO NOT need to be defined explicitly in your desired state file (DSF). However, if you would like to configure some repo with the name stable for example, you can override the default repo.

This example would have `stable` and `incubator` added by default and another `custom` repo defined explicitly:

```toml

 
[helmRepos]
  custom = "https://mycustomrepo.org"

```

```yaml

helmRepos:
  custom: "https://mycustomrepo.org"


```

This example would have `stable` overriden with a custom repo:

```toml
...

[helmRepos]
stable = "https://mycustomstablerepo.com"
...

```

```yaml
...

helmRepos:
  stable: "https://mycustomstablerepo.com"
...

```
