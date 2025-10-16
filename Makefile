.PHONY: build test testacc install lint fmt clean help

BINARY_NAME=terraform-provider-cyberark-sia
VERSION?=dev
INSTALL_PATH=~/.terraform.d/plugins/local/aaearon/cyberark-sia/$(VERSION)/linux_amd64

# Default target
help:
	@echo "Available targets:"
	@echo "  build     - Build the provider binary"
	@echo "  test      - Run unit tests"
	@echo "  testacc   - Run acceptance tests (requires TF_ACC=1)"
	@echo "  install   - Install provider locally for development"
	@echo "  lint      - Run golangci-lint"
	@echo "  fmt       - Format Go code"
	@echo "  clean     - Clean build artifacts"

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
