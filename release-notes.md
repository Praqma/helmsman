# v3.4.5

If you migrating from Helmsman v1.x, it is recommended you read the [migration guide](https://github.com/Praqma/helmsman/blob/master/docs/how_to/misc/migrate_to_3.md) before using this release.

> Starting from Helmsman v3.0.0 GA release, support for Helmsman v1.x will be limited to bug fixes.

# Fixes and improvements:

- Added `--force-update` to the `helm repo add` command to support helm v3.3.3+
- Removed duplicated user and password when adding helm repos.
- Added missing packages to the final docker image.
- Updated some dependencies versions.
- Helm tests are now execurted as the last command after the hooks.

# New features:

- Added post-render option.

None.
