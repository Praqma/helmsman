# v3.5.0

If you migrating from Helmsman v1.x, it is recommended you read the [migration guide](https://github.com/Praqma/helmsman/blob/master/docs/how_to/misc/migrate_to_3.md) before using this release.

> Starting from Helmsman v3.0.0 GA release, support for Helmsman v1.x will be limited to bug fixes.

# Fixes and improvements:

- Handle default app name even when --skip-validation is used; PR #526
- extractChartName works now even when only dev versions are published; PR #532
- Remove references to the deprecated stable repository; PR #533
- Fix no labels on first failed deploy with running release labeling as defer before release exec; PR #541

# New features:

- Add --always-upgrade flag; PR #534
- Add retryExec func for getting cluster status methods; PR #540
