# v3.6.3

If you migrating from Helmsman v1.x, it is recommended you read the [migration guide](https://github.com/Praqma/helmsman/blob/master/docs/how_to/misc/migrate_to_3.md) before using this release.

> Starting from Helmsman v3.0.0 GA release, support for Helmsman v1.x will be limited to bug fixes.

## Fixes and improvements

- Fixed missing diff on chart version change (#557)
- Fixed checking for updates on disabled releases (#557)
- Fixed segmentation fault on slack notifications (#559)
- Fixed failure to remove untracked releases (#566)
- The debug flag is now passed down to the helm commands (#568)
- Improved error reporting (#568)
