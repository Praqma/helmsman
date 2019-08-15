# v1.12.0

> If you are already using an older version of Helmsman than v1.4.0-rc, please read the changes below carefully and follow the upgrade guide [here](docs/migrating_to_v1.4.0-rc.md)

# Fixes and improvements:
- Fix chart version validations and support wildcard `*` versions. PRs #283 #284 

# New features:
- support specifying `--history-max` tiller parameter in the DSF. PR #282
- adding `no-env-values-subst` flag to allow disabling environment variables substitution in values files only.  PR #281 

