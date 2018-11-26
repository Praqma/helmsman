# v1.7.1

> If you are already using an older version of Helmsman than v1.4.0-rc, please read the changes below carefully and follow the upgrade guide [here](docs/migrating_to_v1.4.0-rc.md)

- Improving env variable expansion across DSFs. PR #122. This change requires `$` signs in strings inside DSFs to be escaped as `$$`. If not escaped, they will be treated as variables.   
- Adding `suppress-diff-secrets` for suppressing helm diff secrets. PR #132
- Support overriding existing values when merging DSFs. PR #119
- Fixing minor issues. PRs #131 #128 #130 #131 #118 

