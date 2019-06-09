---
version: v1.6.2
---

# Delete all deployed releases

Helmsman allows you to delete all the helm releases that were deployed by Helmsman from a given desired state.

The `--destroy` flag will remove all deployed releases from a given desired state file (DSF). Note that this does not currently delete the namespaces nor the Kubernetes contexts created.

The deletion of releases will respect the `purge` options in the desired state file. i.e. only if `purge` is true for release A, then the destruction of A will be a purge delete

This was originally requested in issue [#88](https://github.com/Praqma/helmsman/issues/88).

