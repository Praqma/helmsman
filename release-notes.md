# v1.4.0-rc

> If you are already using an older version of Helmsman, please read the changes below carefully and follow the upgrade guide [here](docs/migrating_to_v1.4.0-rc.md)

- Slack notifications for Helmsman plan and execution results. Issue #49
- RBAC Improvements:
    - Validation for service accounts to be used for deploying Tillers. Issue #47
    - Support creating RBAC service accounts for configuring Tiller(s) if they don't exist. Issue 48# 
- Improvements for Multi-tenancy: 
    - Adding `tillerNamespace` option for releases to select which Tiller should manage a release. Issue #32
    - Allowing releases with the same name to be deployed with different Tillers. Issue #50
    - Tracking Helmsman managed releases with special Helmsman labels.
- Adding `--apply-labels` flag to label releases defined in the desired state file with Helmsman's labels.
- Making the name option for Apps optional and using the app name from the (toml/yaml) desired state as a release name when this option is not set.
- Changing Helmsman behavior when removing/commenting out a release in the Apps section of the desired state. Removing/commenting out a release in the desired state will result in **deleting** the release if it's labeled as `managed-by Helmsman`.
    