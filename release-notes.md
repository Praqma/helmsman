# v3.1.0

This is a minor release.
It is recommended you read the [Helm 3 migration guide](https://helm.sh/docs/topics/v2_v3_migration/) before using this release.

> Starting from Helmsman v3.0.0 GA release, support for Helmsman v1.x will be limited to bug fixes.

# Fixes and improvements:
- Add helm v3.0.3 to Docker images built
- Fix multiple versions of chart found in the helm repositories; PR #397
- Get existing helm repos first before adding new ones from DSF in order to limit actions to be taken; PR #403
- Enhance the way -target flag checks releases states; PR #405
- Take advantage of enhancements in -target flag flow for -group flags; PR #407

# New features:
None
