include prostack-go/make/consts.mk

##
## Tests:
.PHONY: test
test: ## - Run tests for all services
test: test-app1

.PHONY: test-app1
test-app1: ## - Run tests for app1
	@echo "Running tests for app1..."
	@cd apps/app1 && go test -v ./...

include prostack-go/make/help.mk
