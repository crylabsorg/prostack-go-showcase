# Entry point for a system assembled from Go apps in apps/.
ROOT_DIR := $(abspath $(dir $(firstword $(MAKEFILE_LIST))))

include $(ROOT_DIR)/prostack-go/make/consts.mk
include $(ROOT_DIR)/prostack-go/make/apps.mk
include $(ROOT_DIR)/prostack-go/make/migrations.mk
include $(ROOT_DIR)/prostack-go/make/proto.mk
include $(ROOT_DIR)/prostack-go/make/compose.mk

OAPI_CODEGEN_VERSION := v2.8.0

## Swagger:
.PHONY: install-oapi-codegen
install-oapi-codegen: ## Install pinned oapi-codegen in bin/ if needed
	@if [ ! -x "$(ROOT_DIR)/bin/oapi-codegen" ] || \
		[ "$$(go version -m "$(ROOT_DIR)/bin/oapi-codegen" | \
			awk '$$1 == "mod" && $$2 == "github.com/oapi-codegen/oapi-codegen/v2" { print $$3 }')" != "$(OAPI_CODEGEN_VERSION)" ]; then \
		GOBIN="$(ROOT_DIR)/bin" go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION); \
	fi

.PHONY: build-gateway-swagger
build-gateway-swagger: install-oapi-codegen ## Generate gateway models and HTTP routes from OpenAPI
	@set -e; cd "$(APPS_DIR)/gateway"; for config in oapi/*.yaml; do \
		"$(ROOT_DIR)/bin/oapi-codegen" --config "$$config" swagger.json; \
	done

MODULE_DIRS := $(APP_DIRS) $(ROOT_DIR)
include $(ROOT_DIR)/prostack-go/make/modules.mk
include $(ROOT_DIR)/prostack-go/make/lint.mk
include $(ROOT_DIR)/prostack-go/make/help.mk
