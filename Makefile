.DEFAULT_GOAL := help

PKGS   := $(shell go list ./... | grep -v /vendor/)
TAG    := $(shell git describe --always --tags --abbrev=0 HEAD)
LAST   := $(shell git describe --always --tags --abbrev=0 HEAD^)
BODY   := "`git log ${LAST}..HEAD --oneline --decorate` `printf '\n\#\#\# [Build Info](${BUILD_URL})'`"
DATE   := $(shell date +'%d%m%y')

# Ensure we have an unambiguous GOPATH.
GOPATH := $(shell go env GOPATH)

ifneq "$(or $(findstring :,$(GOPATH)),$(findstring ;,$(GOPATH)))" ""
  $(error GOPATHs with multiple entries are not supported)
endif

GOPATH := $(realpath $(GOPATH))
ifeq ($(strip $(GOPATH)),)
  $(error GOPATH is not set and could not be automatically determined)
endif

SRCDIR := $(GOPATH)/src/

ifeq ($(filter $(GOPATH)%,$(CURDIR)),)
  GOPATH := $(shell mktemp -d "/tmp/dep.XXXXXXXX")
  SRCDIR := $(GOPATH)/src/
endif

ifneq ($(OS),Windows_NT)
  # Before we start test that we have the mandatory executables available
  EXECUTABLES = go
  OK := $(foreach exec,$(EXECUTABLES),\
    $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH, please install $(exec)")))
endif

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
	@go fmt
.PHONY: fmt

dependencies: ## Ensure all the necessary dependencies
	@go get -t -d -v ./...
.PHONY: dependencies

$(SRCDIR):
	@mkdir -p $(SRCDIR)
	@ln -s $(CURDIR) $(SRCDIR)

dep: $(SRCDIR) ## Ensure vendors with dep
	@cd $(SRCDIR)helmsman && \
	  dep ensure
.PHONY: dep

dep-update: $(SRCDIR) ## Ensure vendors with dep
	@cd $(SRCDIR)helmsman && \
	  dep ensure --update
.PHONY: dep-update

build: dep ## Build the package
	@cd $(SRCDIR)helmsman && \
	  go build -ldflags '-X main.version="${TAG}-${DATE}" -extldflags "-static"'

generate:
	@go generate #${PKGS}
.PHONY: generate

check: $(SRCDIR)
	@cd $(SRCDIR)helmsman && \
	  dep check && \
	  go vet #${PKGS}
.PHONY: check

test: dep ## Run unit tests
	@cd $(SRCDIR)helmsman && \
	  go test -v -cover -p=1 -args -f example.toml
.PHONY: test

cross: dep ## Create binaries for all OSs
	@cd $(SRCDIR)helmsman && \
	  env CGO_ENABLED=0 gox -os '!freebsd !netbsd' -arch '!arm' -output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}" -ldflags '-X main.Version=${TAG}-${DATE}'
.PHONY: cross

release: dep ## Generate a new release
	@cd $(SRCDIR)helmsman && \
	  goreleaser --release-notes release-notes.md

tools: ## Get extra tools used by this makefile
	@go get -u github.com/golang/dep/cmd/dep
	@go get -u github.com/mitchellh/gox
	@go get -u github.com/goreleaser/goreleaser
.PHONY: tools
