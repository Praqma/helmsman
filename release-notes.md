# v1.5.0

> If you are already using an older version of Helmsman than v1.4.0-rc, please read the changes below carefully and follow the upgrade guide [here](docs/migrating_to_v1.4.0-rc.md)

- Adding `--keep-untracked-releases` option to prevent cleaning up untracked release that are managed by Helmsman.
- Restricting the untracked releases clean up (when enabled) to the ones only in the defined namespace in the desired state file.
- Support using multiple desired state files which are merged at runtime. Issue #62. Thanks to @norwoodj
- Support using the `json` output of newer versions of helm. Fixes #61. Thanks to @luisdavim
- Fix relative paths for values files. Issue #59
- Adding `timeout` & `no-hooks` as additional release deployment options. Issue #55

    