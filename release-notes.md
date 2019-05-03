# v1.9.0

> If you are already using an older version of Helmsman than v1.4.0-rc, please read the changes below carefully and follow the upgrade guide [here](docs/migrating_to_v1.4.0-rc.md)

# Fixes and improvements:
- Cleanup helm secrets if any. PR #232
- Add `autoscaling` apiGroup to thelmsman-created Tiller roles. PR #235
- Set proper stderr and stdout for log streams. PR #239

# New features:
- Add -target flag to limit execution to specific app. PR #237
- Substitute environment variables in helm values/secrets files. PR #240

