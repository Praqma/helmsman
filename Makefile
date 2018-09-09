.DEFAULT_GOAL := help

PKGS := $(shell go list ./... | grep -v /vendor/)
TAG   = $(shell git describe --always --tags --abbrev=0 HEAD)
LAST  = $(shell git describe --always --tags --abbrev=0 HEAD^)
BODY  = "`git log ${LAST}..HEAD --oneline --decorate` `printf '\n\#\#\# [Build Info](${BUILD_URL})'`"
DATE  = $(shell date +'%d%m%y')

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

build: dependencies ## Build the package
	@go build -ldflags '-X main.version="${TAG}-${DATE}" -extldflags "-static"'

generate:
	@go generate #${PKGS}
.PHONY: generate

check:
	@go vet #${PKGS}
.PHONY: check

test: dependencies ## Run unit tests
	@go test -v -cover -p=1 #${PKGS}
.PHONY: test

cross: dependencies ## Create binaries for all OSs
	@env CGO_ENABLED=0 gox -os '!freebsd !netbsd' -arch '!arm' -output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}" -ldflags '-X main.Version=${TAG}-${DATE}'
.PHONY: cross

release: dependencies ## Generate a new release
	goreleaser --release-notes release-notes.md

tools: ## Get extra tools used by this makefile
	@go get -u github.com/mitchellh/gox
	@go get -u github.com/goreleaser/goreleaser
.PHONY: tools
