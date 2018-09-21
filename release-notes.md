# v1.6.0

> If you are already using an older version of Helmsman than v1.4.0-rc, please read the changes below carefully and follow the upgrade guide [here](docs/migrating_to_v1.4.0-rc.md)

- Adding support for helm secrets plugin (thanks to @luisdavim). issue #54
- Adding support for using helm diff to determine when upgrades are needed and show the diff.
- Allowing env vars to be loaded from files using Godotenv (thanks to @luisdavim)
- Adding `--dry-run` option in Helmsman to perform a helm dry-run operation. Issues #77 #60
- Adding `useTiller` option in namespaces definitions to use existing Tillers. Issue #71
- Other minor code improvements and color coded output.

