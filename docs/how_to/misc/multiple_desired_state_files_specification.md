---
version: v3.8.1
---

# Specification file

Starting from v3.8.0, Helmsman allows you to use Specification file passed with `--spec <file>` flag
in order to define multiple Desired State Files to be merged in particular order and with specific priorities.

An example Specification file `spec.yaml`:

```yaml
---
stateFiles:
  - path: examples/example.yaml
  - path: examples/minimal-example.yaml
    priority: -10
  - path: examples/minimal-example.toml
    priority: -20

```

This file can be then run with:

```shell
helmsman --spec spec.yaml ...
```

What it does is it takes the files from `stateFiles` list and orders them based on their priorities same way it does with the apps in DSF file.
In an example above the result order would be:

```yaml
  - path: examples/minimal-example.toml
  - path: examples/minimal-example.yaml
  - path: examples/example.yaml
```

with priorities being `-20, -10, 0` after ordering.

Once ordering is done, Helmsman will read each file one by one and merge the previous states with the current file it goes through.

One can take advantage of that and define the state of the environment starting with more general definitions and then reaching more specific cases in the end,
which would overwrite or extend things from previous files.
