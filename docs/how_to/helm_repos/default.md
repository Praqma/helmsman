---
version: v3.0.0-beta5
---

# Default helm repos

Helm v3 no longer adds the `stable` and `incubator` repos by default. Up to Helmsman v3.0.0-beta5, Helmsman adds these two repos by default. And you can disable the automatic addition of these two repos, use the `--no-default-repos` flag.

Starting from `v3.0.0-beta6`, Helmsman complies with the Helm v3 behavior and DOES NOT add `stable` nor `incubator` by default. The `--no-default-repos` is also deprecated.
 

This example would have only the `custom` repo defined explicitly:

```toml


[helmRepos]
  custom = "https://mycustomrepo.org"

```

```yaml

helmRepos:
  custom: "https://mycustomrepo.org"


```

This example would have `stable` defined with a custom repo:

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

This example would have `stable` defined with a Google deprecated stable repo:

```toml
...

[helmRepos]
stable = "https://kubernetes-charts.storage.googleapis.com"
...

```

```yaml
# ...

helmRepos:
  stable: "https://kubernetes-charts.storage.googleapis.com"
# ...

```
