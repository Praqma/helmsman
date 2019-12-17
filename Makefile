.DEFAULT_GOAL := help

PKGS    := $(shell go list ./... | grep -v /vendor/)
TAG     := $(shell git describe --always --tags --abbrev=0 HEAD)
LAST    := $(shell git describe --always --tags --abbrev=0 HEAD^)
BODY    := "`git log ${LAST}..HEAD --oneline --decorate` `printf '\n\#\#\# [Build Info](${BUILD_URL})'`"
DATE    := $(shell date +'%d%m%y')
PRJNAME := $(shell basename "$(PWD)")

# Ensure we have an unambiguous GOPATH.
GOPATH := $(shell go env GOPATH)

ifneq ($(strip $(CIRCLE_WORKING_DIRECTORY)),)
  GOPATH := $(subst /src/$(PRJNAME),,$(CIRCLE_WORKING_DIRECTORY))
  $(info "Using CIRCLE_WORKING_DIRECTORY for GOPATH")
endif

ifneq "$(or $(findstring :,$(GOPATH)),$(findstring ;,$(GOPATH)))" ""
  GOPATH := $(lastword $(subst :, ,$(GOPATH)))
  $(info GOPATHs with multiple entries are not supported, defaulting to the last path in GOPATH)
endif

GOPATH := $(realpath $(GOPATH))
ifeq ($(strip $(GOPATH)),)
  $(error GOPATH is not set and could not be automatically determined)
endif

SRCDIR := $(GOPATH)/src/
PRJDIR := $(CURDIR)

ifeq ($(filter $(GOPATH)%,$(CURDIR)),)
  GOPATH := $(shell mktemp -d "/tmp/dep.XXXXXXXX")
  SRCDIR := $(GOPATH)/src/
  PRJDIR := $(SRCDIR)$(PRJNAME)
endif

ifneq ($(OS),Windows_NT)
  # Before we start test that we have the mandatory executables available
  EXECUTABLES = go
  OK := $(foreach exec,$(EXECUTABLES),\
    $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH, please install $(exec)")))
endif

export CGO_ENABLED=0

help:
	@echo "Available options:"
	@grep -E '^[/1-9a-zA-Z._%-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	   | sort \
	   | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-45s\033[0m %s\n", $$1, $$2}'
.PHONY: help

clean: ## Remove build artifacts
	@git clean -fdX
.PHONY: clean

fmt: ## Reformat package sources
	@go fmt ./...
.PHONY: fmt

dependencies: ## Ensure all the necessary dependencies
	@go get -t -d -v ./...
.PHONY: dependencies

$(SRCDIR):
	@mkdir -p $(SRCDIR)
	@ln -s $(CURDIR) $(SRCDIR)

dep: $(SRCDIR) ## Ensure vendors with dep
	@cd $(PRJDIR) && \
	  dep ensure -v
.PHONY: dep

dep-update: $(SRCDIR) ## Ensure vendors with dep
	@cd $(PRJDIR) && \
	  dep ensure -v --update
.PHONY: dep-update

build: dep ## Build the package
	@cd $(PRJDIR) && \
	  go build -o helmsman -ldflags '-X main.version="${TAG}-${DATE}" -extldflags "-static"' cmd/helmsman/main.go

generate:
	@go generate #${PKGS}
.PHONY: generate

check: $(SRCDIR) fmt
	@cd $(PRJDIR) && \
	  dep check && \
	  go vet ./...
.PHONY: check

repo:
	@cd $(PRJDIR) && \
		helm repo add stable https://kubernetes-charts.storage.googleapis.com
.PHONY: repo

test: dep check repo ## Run unit tests
	@go test -v -cover -p=1 ./... -args -f ../../examples/example.toml
.PHONY: test

cross: dep ## Create binaries for all OSs
	@cd $(PRJDIR) && \
	  gox -os '!freebsd !netbsd' -arch '!arm' -output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}" -ldflags '-X main.Version=${TAG}-${DATE}' ./...
.PHONY: cross

release: ## Generate a new release
	@cd $(PRJDIR) && \
	  goreleaser --release-notes release-notes.md --rm-dist

tools: ## Get extra tools used by this makefile
	@go get -u github.com/golang/dep/cmd/dep
	@go get -u github.com/mitchellh/gox
	@go get -u github.com/goreleaser/goreleaser
.PHONY: tools

helmPlugins: ## Install helm plugins used by Helmsman
	@mkdir -p ~/.helm/plugins
	@helm plugin install https://github.com/hypnoglow/helm-s3.git
	@helm plugin install https://github.com/nouney/helm-gcs
	@helm plugin install https://github.com/databus23/helm-diff
	@helm plugin install https://github.com/futuresimple/helm-secrets
	@helm plugin install https://github.com/rimusz/helm-tiller
.PHONY: helmPlugins
