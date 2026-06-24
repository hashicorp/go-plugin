default: lint test

.PHONY: deps
deps:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@c0d3ddc9cf3faa61a4e378e879ece580256d76e5 # v2.12.2

.PHONY: lint
lint:
	@echo "==> Running linters..."
	@golangci-lint run
	@echo "==> Done"

.PHONY: test-deps
test-deps:
	@echo "==> Building test fixtures..."
	@$(MAKE) -C internal/cmdrunner/testdata
	@echo "==> Done"

.PHONY: test
test: test-deps
	@echo "==> Running tests..."
	@go test -v -race -timeout=60s ./...
	@echo "==> Done"
