# v3.0.2

This is a bugfix release to support Helm v3.
It is recommended you read the [Helm 3 migration guide](https://helm.sh/docs/topics/v2_v3_migration/) before using this release.

> Starting from Helmsman v3.0.0 GA release, support for Helmsman v1.x will be limited to bug fixes.

# Fixes and improvements:
- Add concurrency limit on goroutines for getting release state and decision makers; PR #395
- Add second 'Updated' time layout to parser for release state; PR #393
- Fix unmarshal issue for quoted version in Chart.yaml; PR #389

# New features:
None, bug fix release
