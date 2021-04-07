# v3.6.7

If you migrating from Helmsman v1.x, it is recommended you read the [migration guide](https://github.com/Praqma/helmsman/blob/master/docs/how_to/misc/migrate_to_3.md) before using this release.

> Starting from Helmsman v3.0.0 GA release, support for Helmsman v1.x will be limited to bug fixes.

## Fixes and improvements


- Updated the Docker images to use the latest helm version (#597)
- Replaced the deprecated helm-secrets repo with the new one (#595)
- Fised issue preventing the proper expansion of paths for executable hooks (#594)
- Helmsman will now skip helm tests in dry-run mode (#593)
- Dont expect a username and password if the caClient cert is present (#592)
- Added flag to skip helm repo updates (#590)
