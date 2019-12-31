---
version: v3.0.0-beta4
---

# Default helm repos

Helm v3 no longer adds the `stable` and `incubator` repos by default. However, Helmsman v3.0.0-beta4 still adds these two repos by default. These two DO NOT need to be defined explicitly in your desired state file (DSF). However, if you would like to configure some repo with the name stable for example, you can override the default repo.

> You can disable the automatic addition of these two repos, use the `--no-default-repos` flag.

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
# ...

helmRepos:
  stable: "https://mycustomstablerepo.com"
# ...

```
