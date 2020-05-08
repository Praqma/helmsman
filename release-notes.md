# v3.3.0

If you migrating from Helmsman v1.x, it is recommended you read the [migration guide](https://github.com/Praqma/helmsman/blob/master/docs/how_to/misc/migrate_to_3.md) before using this release.

> Starting from Helmsman v3.0.0 GA release, support for Helmsman v1.x will be limited to bug fixes.

# Fixes and improvements:
- Add DRY-ed examples using appsTemplates; PR #425

# New features:
- Add `--p` flag to define parallel apps installation/upgrade for those with the same priority defined in DSF; PR #431
- Add Namespace's resource quotas to DSF; PR #384
