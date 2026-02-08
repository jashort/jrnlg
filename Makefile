.PHONY: all build test clean install uninstall run help coverage fmt lint vet

# Variables
BINARY_NAME=jrnlg
VERSION?=0.1.0
BUILD_DIR=build
INSTALL_PATH=/usr/local/bin
GO=go

# Linker flags to inject version information
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

all: build ## Build the binary

help: ## Show this help message
	@echo "jrnlg Makefile"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Multi-platform build complete in $(BUILD_DIR)/"

test: ## Run all tests
	@echo "Running tests..."
	$(GO) test -v ./...
	@echo "All tests passed"

test-short: ## Run tests without verbose output
	@echo "Running tests..."
	$(GO) test ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GO) test -cover ./...
	@echo ""
	@echo "Detailed coverage:"
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

coverage: test-coverage ## Alias for test-coverage

coverage-html: ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

fmt: ## Format code with gofmt
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "No issues found"

lint: ## Run golangci-lint
	@echo "Running linter..."
	go tool github.com/golangci/golangci-lint/cmd/golangci-lint run ./...; \
	echo "Linting complete"; \

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

install: build ## Install binary to system
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@sudo mv $(BINARY_NAME) $(INSTALL_PATH)/
	@sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installed successfully"
	@echo "Run '$(BINARY_NAME) --help' to get started"

uninstall: ## Uninstall binary from system
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_PATH)..."
	@sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstalled successfully"

deps: ## Download dependencies
	@echo "Downloading dependencies...$(NC)"
	$(GO) mod download
	@echo "✓ Dependencies downloaded$(NC)"

deps-tidy: ## Tidy dependencies
	@echo "Tidying dependencies...$(NC)"
	$(GO) mod tidy
	@echo "✓ Dependencies tidied$(NC)"

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies...$(NC)"
	$(GO) mod verify
	@echo "✓ Dependencies verified$(NC)"

check: fmt vet test ## Run fmt, vet, and test

ci: deps-verify fmt vet test ## Run all CI checks

release: clean check build-all ## Prepare a release (clean, check, build-all)
	@echo ""
	@echo "Release build complete!"
	@echo "Built binaries:"
	@ls -lh $(BUILD_DIR)/
