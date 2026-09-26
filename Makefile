# Entry point for a system assembled from Go apps in apps/.
ROOT_DIR := $(abspath $(dir $(firstword $(MAKEFILE_LIST))))

include $(ROOT_DIR)/prostack-go/make/consts.mk
include $(ROOT_DIR)/prostack-go/make/apps.mk
include $(ROOT_DIR)/prostack-go/make/migrations.mk
include $(ROOT_DIR)/prostack-go/make/proto.mk
include $(ROOT_DIR)/prostack-go/make/compose.mk

MODULE_DIRS := $(APP_DIRS) $(ROOT_DIR)
include $(ROOT_DIR)/prostack-go/make/modules.mk
include $(ROOT_DIR)/prostack-go/make/help.mk
