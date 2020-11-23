# v3.6.0

If you migrating from Helmsman v1.x, it is recommended you read the [migration guide](https://github.com/Praqma/helmsman/blob/master/docs/how_to/misc/migrate_to_3.md) before using this release.

> Starting from Helmsman v3.0.0 GA release, support for Helmsman v1.x will be limited to bug fixes.

## New features

- Lifecycle hooks can now be shell commands or script in addition to k8s resources (#543)
- The Helm release secrets are annotated with the result from the lifecycle hooks

## Fixes and improvements

- Code cleanup
- Performance improvements (#543; #545; #547)
- Fixed a bug when the chart is renamed (#546)
