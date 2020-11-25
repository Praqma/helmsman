# v3.6.2

If you migrating from Helmsman v1.x, it is recommended you read the [migration guide](https://github.com/Praqma/helmsman/blob/master/docs/how_to/misc/migrate_to_3.md) before using this release.

> Starting from Helmsman v3.0.0 GA release, support for Helmsman v1.x will be limited to bug fixes.

## Fixes and improvements

- Commands will optionally (Verbose) log stdout on non-zero exit code (#554)
- `helm test` command will run immediately after install / upgrade (before hooks and labelling). (#553)
- Add test cases for `readState`; improved error handling and small order-of-operations refactor. (#552)
