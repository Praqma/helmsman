# Migrating to Helmsman v1.4.0-rc or higher

This document highlights the main changes between Helmsman v1.4.0-rc and the older versions. While the changes are still backward-compatible, the behavior and the internals have changed. The list below highlights those changes:

- Helmsman v1.4.0-rc tracks the releases it manages by applying specific labels to their Helm state (stored in Helm's configured backend storage). For smooth transition when upgrading to v1.4.0-rc, you should run `helmsman -f <your desired state file> --apply-labels` once. This will label all releases from your desired state as with a `MANAGED-BY=Helmsman` label. The `--apply-labels`is safe to run multiple times.

- After each run, Helmsman v1.4.0-rc looks for, and deletes any releases with the `MANAGED-BY=Helmsman` label which are no longer existing in your desired state. This means that **deleting/commenting out an app from your desired state file will result in its deletion**. You can disable this cleanup by adding the flag `--keep-untracked-releases` to your Helmsman commands.

