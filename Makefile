# Entry point for a system assembled from Go apps in apps/.
ROOT_DIR := $(abspath $(dir $(firstword $(MAKEFILE_LIST))))

include $(ROOT_DIR)/prostack-go/make/consts.mk
include $(ROOT_DIR)/prostack-go/make/apps.mk
include $(ROOT_DIR)/prostack-go/make/migrations.mk
include $(ROOT_DIR)/prostack-go/make/proto.mk
include $(ROOT_DIR)/prostack-go/make/swagger.mk
include $(ROOT_DIR)/prostack-go/make/compose.mk

## Swagger:
.PHONY: build-gateway-swagger
build-gateway-swagger: install-swag ## Generate gateway OpenAPI 3.1 documentation
	@cd "$(APPS_DIR)/gateway" && "$(ROOT_DIR)/bin/swag" init \
		--generalInfo main.go \
		--dir "./,$(ROOT_DIR)/prostack-go/controller" \
		--output "$(APPS_DIR)/gateway/internal/swagger" \
		--outputTypes json \
		--v3.1

MODULE_DIRS := $(APP_DIRS) $(ROOT_DIR)
include $(ROOT_DIR)/prostack-go/make/modules.mk
include $(ROOT_DIR)/prostack-go/make/lint.mk
include $(ROOT_DIR)/prostack-go/make/help.mk
