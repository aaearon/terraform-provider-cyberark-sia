.PHONY: build test testacc install lint fmt clean help check-env test-crud deps generate

BINARY_NAME=terraform-provider-cyberark-sia
VERSION?=dev
INSTALL_PATH=~/.terraform.d/plugins/local/aaearon/cyberark-sia/$(VERSION)/linux_amd64

# Default target
help:
	@echo "Available targets:"
	@echo "  build       - Build the provider binary"
	@echo "  test        - Run unit tests"
	@echo "  testacc     - Run acceptance tests (requires TF_ACC=1)"
	@echo "  install     - Install provider locally for development"
	@echo "  lint        - Run golangci-lint"
	@echo "  fmt         - Format Go code"
	@echo "  clean       - Clean build artifacts"
	@echo "  check-env   - Verify required environment variables are set"
	@echo "  test-crud   - Run automated CRUD validation (usage: make test-crud DESC=resource-name)"
	@echo "  deps        - Download and tidy Go dependencies"
	@echo "  generate    - Generate provider documentation"

# Build the provider binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME)

# Run unit tests
test:
	@echo "Running unit tests..."
	go test -v -race -timeout=30s ./...

# Run acceptance tests (requires TF_ACC=1)
testacc:
	@echo "Running acceptance tests..."
	TF_ACC=1 go test -v -race -timeout=120m ./internal/provider

# Install provider locally for Terraform development
install: build
	@echo "Installing provider to $(INSTALL_PATH)..."
	@mkdir -p $(INSTALL_PATH)
	@cp $(BINARY_NAME) $(INSTALL_PATH)/

# Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

# Format Go code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	gofmt -s -w .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -rf vendor/
	@go clean -cache

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Generate documentation
generate:
	@echo "Generating documentation..."
	go generate ./...

# Check required environment variables
check-env:
	@echo "Checking required environment variables..."
	@test -n "$(CYBERARK_USERNAME)" || (echo "❌ CYBERARK_USERNAME not set (see CLAUDE.md → Environment Setup)" && exit 1)
	@test -n "$(CYBERARK_CLIENT_SECRET)" || (echo "❌ CYBERARK_CLIENT_SECRET not set (see CLAUDE.md → Environment Setup)" && exit 1)
	@echo "✅ Required environment variables are set"
	@if [ -z "$(TF_ACC)" ]; then \
		echo "⚠️  TF_ACC not set (recommended: export TF_ACC=1)"; \
	fi

# Run automated CRUD testing workflow
test-crud: check-env
	@if [ -z "$(DESC)" ]; then \
		echo "Usage: make test-crud DESC=<resource-description>"; \
		echo "Example: make test-crud DESC=policy-principal-assignment"; \
		exit 1; \
	fi
	@echo "Running CRUD validation for: $(DESC)"
	./scripts/test-crud-resource.sh "$(DESC)"
