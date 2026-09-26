GOLANGCI_LINT_VERSION := $(shell cat .golangci-lint-version)
GOLANGCI_LINT := .bin/$(GOLANGCI_LINT_VERSION)/golangci-lint

.PHONY: lint

lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...

$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh | sh -s -- -b .bin/$(GOLANGCI_LINT_VERSION) $(GOLANGCI_LINT_VERSION)
