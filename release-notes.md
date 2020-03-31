# v3.2.0

If you migrating from Helmsman v1.x, it is recommended you read the [migration guide](https://github.com/Praqma/helmsman/blob/master/docs/how_to/misc/migrate_to_3.md) before using this release.

> Starting from Helmsman v3.0.0 GA release, support for Helmsman v1.x will be limited to bug fixes.

# Fixes and improvements:
- Disable checking helm-secrets plugin installed when hiera-eyaml used for secrets; PR #408
- Make decide exiting on pending-upgrade chart state; PR #409
- Changes to standard output and Slack messages; PR #410
- Fix local charts installation not being idempotent; PR #411
- Fix destroy command when repo cache is not up to date; PR #416
- Fix namespace limits defined in TOML not properly parsed; PR #422
- Fix race condition when applying limit ranges to multiple namespaces; PR #429


# New features:
- Add `--context-override` flag to skip fetching context name from releases state; PR #420
- Add `no-cleanup` flag to optionally keep k8s credentials on host; PR #423
- Add `--migrate-context` flag support renaming helmsman contexts and enable smooth migration of helm 2 releases to 3; PR #427
