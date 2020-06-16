# v3.4.4

If you migrating from Helmsman v1.x, it is recommended you read the [migration guide](https://github.com/Praqma/helmsman/blob/master/docs/how_to/misc/migrate_to_3.md) before using this release.

> Starting from Helmsman v3.0.0 GA release, support for Helmsman v1.x will be limited to bug fixes.

# Fixes and improvements:
- Fix missing valid parallel run for cmd exec after hooks were introduced; PR #491
- Fix issue when nil is passed to execCommand for deletion of an untracked release; PR #489

# New features:
None.
