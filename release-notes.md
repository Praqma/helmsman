# v1.8.0

> If you are already using an older version of Helmsman than v1.4.0-rc, please read the changes below carefully and follow the upgrade guide [here](docs/migrating_to_v1.4.0-rc.md)

# New features: 
- Support for connection to clusters with bearer token and therefore allowing helmsman to run inside a k8s cluster. PR #206
- Improved cluster connection validation and error messages. PR #206
- Making the `settings` and `helmRepos` stanzas optional. PR #206 #2014
- Support for private repos with basic auth. PR #214
- Support for using pre-configured helm repos. PR #201
- Support for using Azure blob storage for certificates and keys. PR #213

# Fixes:
- Minor bug fixes: PRs #205 #212

