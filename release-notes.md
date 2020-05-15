# v3.4.1

If you migrating from Helmsman v1.x, it is recommended you read the [migration guide](https://github.com/Praqma/helmsman/blob/master/docs/how_to/misc/migrate_to_3.md) before using this release.

> Starting from Helmsman v3.0.0 GA release, support for Helmsman v1.x will be limited to bug fixes.

# Fixes and improvements:
- Validate env vars used in DSF (and other) files are set in th environment before expanding them. PR #463
- Pass chart version to helm show commands to avoid errors with develop chart versions. PR #464
- Report the right error messages when releases are in pending state. PR #465
- Randomize Helmsman's temporary file names to avoid errors when using the same file basename multiple times. PR #470

# New features:
None.
