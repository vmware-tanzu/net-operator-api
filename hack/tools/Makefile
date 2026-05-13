# If you update this file, please follow
# https://suva.sh/posts/well-documented-makefiles

# Ensure Make is run with bash shell as some syntax below is bash-specific
SHELL := /usr/bin/env bash

.DEFAULT_GOAL := help

# Active module mode, as we use go modules to manage dependencies
export GO111MODULE := on

# Directories.
BIN_DIR := bin

export GOBIN := $(shell git rev-parse --show-toplevel)/hack/tools/bin

HOST_OS=$(shell go env GOOS)
HOST_ARCH=$(shell go env GOARCH)

# Binaries.
CONTROLLER_GEN      := $(BIN_DIR)/controller-gen
GOLANGCI_LINT       := $(BIN_DIR)/golangci-lint
GOLANGCI_LINT_KAL   := $(BIN_DIR)/golangci-lint-kal
CLIENT_GEN          := $(BIN_DIR)/client-gen
INFORMER_GEN        := $(BIN_DIR)/informer-gen
LISTER_GEN          := $(BIN_DIR)/lister-gen

## --------------------------------------
##@ Help
## --------------------------------------

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
##@ Binaries
## --------------------------------------

.PHONY: $(CONTROLLER_GEN)
controller-gen: $(CONTROLLER_GEN) ## Install controller-gen
$(CONTROLLER_GEN): go.mod
	go build -tags=tools -o $@ sigs.k8s.io/controller-tools/cmd/controller-gen

.PHONY: $(CLIENT_GEN)
client-gen: $(CLIENT_GEN) ## Install client-gen
$(CLIENT_GEN): go.mod
	go build -tags=tools -o $@ k8s.io/code-generator/cmd/client-gen

.PHONY: $(INFORMER_GEN)
informer-gen: $(INFORMER_GEN) ## Install informer-gen
$(INFORMER_GEN): go.mod
	go build -tags=tools -o $@ k8s.io/code-generator/cmd/informer-gen

.PHONY: $(LISTER_GEN)
lister-gen: $(LISTER_GEN) ## Install lister-gen
$(LISTER_GEN): go.mod
	go build -tags=tools -o $@ k8s.io/code-generator/cmd/lister-gen

.PHONY: $(GOLANGCI_LINT)
golangci-lint: $(GOLANGCI_LINT) ## Install golangci-lint
# golangci-lint v2+ uses module path .../golangci-lint/v2/cmd/golangci-lint (v1 path no longer valid with Go 1.24+).
GOLANGCI_LINT_VERSION ?= v2.11.4
$(GOLANGCI_LINT):
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

# golangci-lint-kal bundles kube-api-linter (Go 1.24+; see https://github.com/kubernetes-sigs/kube-api-linter).
.PHONY: $(GOLANGCI_LINT_KAL)
golangci-lint-kal: $(GOLANGCI_LINT_KAL) ## Build golangci-lint with kube-api-linter
$(GOLANGCI_LINT_KAL): go.mod go.sum tools.go
	@mkdir -p "$(@D)"
	go build -trimpath -o $@ sigs.k8s.io/kube-api-linter/cmd/golangci-lint-kube-api-linter

## --------------------------------------
##@ Generate
## --------------------------------------

.PHONY: modules
modules: ## Runs go mod tidy to validate modules
	go mod tidy -v

.PHONY: modules-download
modules-download: ## Downloads and caches the modules
	go mod download

## --------------------------------------
##@ Cleanup
## --------------------------------------

.PHONY: clean
clean: ## Run all the clean targets
	$(MAKE) clean-bin

.PHONY: clean-bin
clean-bin: ## Remove all generated binaries
	rm -rf bin
