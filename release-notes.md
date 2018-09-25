# v1.6.1

> If you are already using an older version of Helmsman than v1.4.0-rc, please read the changes below carefully and follow the upgrade guide [here](docs/migrating_to_v1.4.0-rc.md)

- Fixing cluster-wide access problem when checking for untracked releases in restricted clusters. Issue #83.
- Fixing not including releases from existing Tillers in helm state when using useTiller.
- Adding `--no-fancy` option to disable colored output.

