# v3.4.0

If you migrating from Helmsman v1.x, it is recommended you read the [migration guide](https://github.com/Praqma/helmsman/blob/master/docs/how_to/misc/migrate_to_3.md) before using this release.

> Starting from Helmsman v3.0.0 GA release, support for Helmsman v1.x will be limited to bug fixes.

# Fixes and improvements:
- Respect dry-run flag with kubectl commands; PR #462
- Add gnupg to Docker image;  PR #449
- Allow absolute paths for values/secrets/etc files; PR #459
- Fix inconsistent helm args between install and upgrade; PR #458

# New features:
- Add lifecycle hooks into Helmsman; PR #421
- Support for using --history-max helm upgrade flag; PR #460
- Support the Helm --set-file flag; PR #444
