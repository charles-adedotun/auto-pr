# Auto PR - Development Makefile

.PHONY: lint test format build check install-hooks clean help

# Default Go command
GO := go
GOLANGCI_LINT := golangci-lint

# Binary name and output directory
BINARY_NAME := auto-pr
BUILD_DIR := ./bin

# Default target
help: ## Show this help message
	@echo 'Auto PR Development Commands:'
	@echo ''
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'
	@echo ''

check: format lint test ## Run all quality checks (format, lint, test)
	@echo "✅ All quality checks passed!"

format: ## Format code using go fmt
	@echo "🔧 Formatting code..."
	@$(GO) fmt ./...

lint: ## Run linter using golangci-lint
	@echo "🔍 Running linter..."
	@which $(GOLANGCI_LINT) > /dev/null || (echo "❌ golangci-lint not found. Install with: brew install golangci-lint" && exit 1)
	@$(GOLANGCI_LINT) run

test: ## Run all tests
	@echo "🧪 Running tests..."
	@$(GO) test ./... -v

test-race: ## Run tests with race detector
	@echo "🏁 Running tests with race detector..."
	@$(GO) test -race ./...

build: format ## Build the binary
	@echo "🏗️  Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✅ Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

build-release: ## Build release binaries for multiple platforms
	@echo "🚀 Building release binaries..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@GOOS=windows GOARCH=amd64 $(GO) build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "✅ Release binaries built in $(BUILD_DIR)/"

install-hooks: ## Install git pre-commit hooks
	@echo "🪝 Installing git pre-commit hooks..."
	@mkdir -p .git/hooks
	@echo '#!/bin/bash' > .git/hooks/pre-commit
	@echo 'set -e' >> .git/hooks/pre-commit
	@echo 'echo "🔍 Running pre-commit checks..."' >> .git/hooks/pre-commit
	@echo 'make check' >> .git/hooks/pre-commit
	@echo 'echo "✅ Pre-commit checks passed!"' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Pre-commit hooks installed!"
	@echo "   Run 'git config --bool core.hooksPath .git/hooks' if hooks don't trigger"

vet: ## Run go vet
	@echo "🔬 Running go vet..."
	@$(GO) vet ./...

clean: ## Clean build artifacts and test cache
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@$(GO) clean
	@$(GO) clean -testcache
	@echo "✅ Clean complete"

deps: ## Download and tidy dependencies
	@echo "📦 Managing dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "✅ Dependencies updated"

dev-setup: install-hooks deps ## Setup development environment
	@echo "🛠️  Setting up development environment..."
	@which $(GOLANGCI_LINT) > /dev/null || (echo "Installing golangci-lint..." && brew install golangci-lint)
	@echo "✅ Development environment ready!"
	@echo ""
	@echo "Quick commands:"
	@echo "  make check  - Run all quality checks before committing"
	@echo "  make build  - Build the binary"
	@echo "  make test   - Run tests"

# Development workflow targets
pre-commit: check ## Run pre-commit checks manually
	@echo "✅ Ready to commit!"

# CI simulation
ci: format lint vet test build ## Simulate CI pipeline locally
	@echo "🎯 CI simulation complete!"