# Contribution Guide

Pull requests, feeback/feature requests are all welcome. This guide will be updated overtime. 

## Build helmsman from source

To build helmsman from source, you need go:1.8+.  Follow the steps below:

```
git clone https://github.com/Praqma/helmsman.git
go get github.com/BurntSushi/toml
go get github.com/Praqma/helmsman/gcs 
go get github.com/Praqma/helmsman/aws
TAG=$(git describe --abbrev=0 --tags)-$(date +"%s")
go build -ldflags '-X main.version='$TAG' -extldflags "-static"'
```

## Submitting pull requests

Please make sure you state the purpose of the pull request and that the code you submit is documented. If in doubt, [this guide](https://blog.github.com/2015-01-21-how-to-write-the-perfect-pull-request/) offers some good tips on writing a PR.

## Contribution to documentation

Contribution to the documentation can be done via pull requests or by openeing an issue.

## Reporting issues/featuer requests

Please provide details of the issue, versions of helmsman, helm and kubernetes and all possible logs.

