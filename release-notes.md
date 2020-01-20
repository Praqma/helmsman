# v3.0.1

This is a bugfix release to support Helm v3.
It is recommended you read the [Helm 3 migration guide](https://helm.sh/docs/topics/v2_v3_migration/) before using this release.

> Starting from Helmsman v3.0.0 GA release, support for Helmsman v1.x will be limited to bug fixes.

# Fixes and improvements:
- Fix overlocking issue when looking for untracked releases; PR #378
- Do not use --all-namespaces to find helm releases; PR #379
- Remove --purge from reInstall function; PR #381

# New features:
None, bug fix release
