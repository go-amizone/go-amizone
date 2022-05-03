# Ensure Make is run with bash shell as some syntax below is bash-specific
SHELL := /usr/bin/env bash

# Help by default
.DEFAULT_GOAL:=help

# Load .env file if possible
ifneq (,$(wildcard ./.env))
	include .env
	ENV_LOADED := true
	export
endif

.PHONY: help
help: ## Get help on available targets
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: test
test-unit: ## Run unit tests
	@echo "Running tests..."
	go test -v ./...

.PHONY: test-integration
test-integration: ## Run integration tests
ifneq (true,$(ENV_LOADED))
	@echo "Integration tests require a variables from a .env file. Please create one by copying and populating the dev.env file."
endif
	@echo "Running integration tests..."
	go test -v ./... -tags=integration -run '^\QTestIntegrate'

.PHONY: test-all
test-all: test-unit test-integration ## Run all tests