# v1.13.0

> If you are already using an older version of Helmsman than v1.4.0-rc, please read the changes below carefully and follow the upgrade guide [here](docs/migrating_to_v1.4.0-rc.md)

# Fixes and improvements:
- Fixed issues when runnin tillerless mode. PRs #320 #325 #305 #307
- Improved support for running on Windows. PR #287
- Improved support for local helm charts. PR #314 #291
- Improved build process and moved to go 1.13.1. PRs #306 #313 #317 #322

# New features:
- Support hiera-eyaml as optional solution for secrets encryption. PR #323
- New group flag to allow releasing a subset of apps
- Added support for SSM params in the Helm values file. PR #295
- Added support for yaml anchors. PR #302
- Added flag to opt out of the default Helm repos. PR #304
