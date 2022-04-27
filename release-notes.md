# v3.9.0

## New features

- Added Option for checking for available updates for app charts (#640)
- Added option for waiting for pending helm releases (#646)
- Added `arm64` builds (#642) (#647)

## Fixes and improvements

- Updated dependencies (#641)
- Avoid the extra chart download step for OCI charts (#643)
- Code refactoring (#644)
- Enabled automatic dependency updates through dependabot

## Breaking changes ⚠

- env files loading is now more intuitive (#649)
  - Before the default .env file would only be loaded if no env files were explicitly passed through the -e flag, now it will always be loaded first if present
  - Before loading env files would not overwrite any env variable that had already been set before, now it does so when loading multiple files if a variable is set more than once the value from the last file to be loaded will take precedence.
  - Before the first file would take precedence, now, the last one will.

