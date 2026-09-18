.PHONY: lint

lint:
	GOTOOLCHAIN=auto go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2 run ./...
