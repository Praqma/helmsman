# v1.7.0

> If you are already using an older version of Helmsman than v1.4.0-rc, please read the changes below carefully and follow the upgrade guide [here](docs/migrating_to_v1.4.0-rc.md)

- Resources (Tillers, namespaces, service accounts) will not be created without the `--apply` flag. Fixes #65 and #100.
- adding `--no-ns` option to prevent helmsman from creating namespaces. This allows users to create namespaces using their own custom charts. Fixes #71 and #100.
- Adding namespace labels. PR #111
- Fixing a bug that deletes untracked releases when running with `--dry-run`. PR #110
- Improving values and secrets file pahts resolution relative to dsf(s). PR #109

