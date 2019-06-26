# v1.10.0

> If you are already using an older version of Helmsman than v1.4.0-rc, please read the changes below carefully and follow the upgrade guide [here](docs/migrating_to_v1.4.0-rc.md)

# Fixes and improvements:
- Allow local chart paths to contain white spaces. PR #262
- Add helm flags to helm diff. PR #252 (thanks to @xaka)
- Fix deleting untracked releases if `-target` flag is set. PR #248 (thanks to @fmotrifork)

# New features:
- Add global toggle `--no-env-subst` to be able to disable environment substitution. PR #249 (thanks to @hatemosphere)
- Add `--diff-context` flag to set lines of diff context. PR #251 (thanks to @lachlancooper)
- Allow deploying Tiller with custom Role. PR #258 (thanks to @mkubaczyk)
- Support Tillerless mode on Helm 2 using helm-tiller plugin. PR #259 (thanks to @mkubaczyk and @robbert229)


