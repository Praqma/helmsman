# Contribution Guide

Pull requests, feeback/feature requests are all welcome. This guide will be updated overtime.

## Build helmsman from source

To build helmsman from source, you need go:1.9+.  Follow the steps below:

```
git clone https://github.com/Praqma/helmsman.git
make build
```

## Submitting pull requests

Please make sure you state the purpose of the pull request and that the code you submit is documented. If in doubt, [this guide](https://blog.github.com/2015-01-21-how-to-write-the-perfect-pull-request/) offers some good tips on writing a PR.

## Contribution to documentation

Contribution to the documentation can be done via pull requests or by opening an issue.

## Reporting issues/feature requests

Please provide details of the issue, versions of helmsman, helm and kubernetes and all possible logs.

## Releasing Helmsman

Release is automated from CicrcleCI based on Git tags. [Goreleaser](goreleaser.com) is used to release the binaries and update the release notes on Github while the circleci pipeline builds a set of docker images and pushes them to dockerhub.

The following steps are needed to cut a release (They assume that you are on master and the code is up to date):
1. Change the version variable in [main.go](main.go)
2. Update the [release-notes.md](release-notes.md) file with new version and changelog.
3. (Optional), if new helm versions are required, update the [circleci config](.circleci/config.yml) and add more docker commands.
4. Commit your changes locally. 
5. Create a git tag with the following command: `git tag -a <semantic version number> -m "<semantic version number>" <your-last-commit-sha>`
6. Push your commit and tag with `git push --follow-tags`
7. This should trigger the [pipeline on circleci](https://circleci.com/gh/Praqma/workflows/helmsman) which eventually releases to Github and dockerhub. 
 
