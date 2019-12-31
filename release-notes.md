# v3.0.0-beta4

This is a major release to support Helm v3.
It is recommended you read the [Helm 3 migration guide](https://helm.sh/docs/topics/v2_v3_migration/) before using this release.

> Starting from this release, support for Helmsman v1.x will be limited to bug fixes.

The following are the most important changes:
- A new and improved logger.
- Restructuring the code.
- Parallelized decision making
- Introducing the `context` stanza to define a context for each DSF. More details [here](docs/misc/merge_desired_state_files).
- Deprecating all the DSF stanzas related to Tiller.
- Deprecating the `purge` option for releases.
- The default value for `storageBackend` is now `secret`.
- The `--suppress-diff-secrets` cmd flag is enabled by default.
- The `--no-env-values-subst` cmd flag is enabled by default.
- The `--no-ssm-values-subst` cmd flag is enabled by default.

