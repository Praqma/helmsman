# v1.11.0

> If you are already using an older version of Helmsman than v1.4.0-rc, please read the changes below carefully and follow the upgrade guide [here](docs/migrating_to_v1.4.0-rc.md)

# Fixes and improvements:
- Consider `--target` when validating helm charts. PR#269
- Fix overriding stable repo with a custom one. PR#271
- Add start and finish logs when deploying charts. PR#266

# New features:
- adding `--force-upgrades` option [ changes default upgrade behaviour]. PR#273
- adding `--update-deps` option for local charts dependency update before use. PR#272


