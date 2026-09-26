# Entry point for a system assembled from Go apps in apps/.
ROOT_DIR := $(abspath $(dir $(firstword $(MAKEFILE_LIST))))

include $(ROOT_DIR)/prostack-go/make/consts.mk
include $(ROOT_DIR)/prostack-go/make/apps.mk

MODULE_DIRS := $(APP_DIRS) $(ROOT_DIR)
include $(ROOT_DIR)/prostack-go/make/modules.mk
include $(ROOT_DIR)/prostack-go/make/help.mk

## Local dependencies:
.PHONY: dev-up
dev-up: ## Start local dependencies
	@docker compose -f devops/docker-compose.yaml up -d

.PHONY: dev-down
dev-down: ## Stop local dependencies
	@docker compose -f devops/docker-compose.yaml down

## Protocol buffers:
.PHONY: proto-tools
proto-tools: ## Install pinned protobuf generators into bin/
	@GOBIN=$(ROOT_DIR)/bin go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
	@GOBIN=$(ROOT_DIR)/bin go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

.PHONY: proto-gen
proto-gen: proto-tools ## Generate shared gRPC code in pkg/proto
	@buf generate

## Database migrations:
.PHONY: migrate-up
migrate-up: ## Apply PostgreSQL migrations with Atlas
	@atlas migrate apply --env local

.PHONY: migrate-status
migrate-status: ## Show PostgreSQL migration status
	@atlas migrate status --env local

.PHONY: migrate-new
migrate-new: ## Create an Atlas migration: make migrate-new NAME=description
	@test -n "$(NAME)" || (echo "NAME is required"; exit 1)
	@atlas migrate new "$(NAME)" --dir file://apps/user/migrations

.PHONY: migrate-hash
migrate-hash: ## Recompute Atlas migration checksum after manual edits
	@atlas migrate hash --dir file://apps/user/migrations
