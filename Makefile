# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary info
BINARY_NAME=exchange
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe

# Version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME = $(shell date +%Y-%m-%dT%H:%M:%S%z)
GIT_COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH = $(shell git branch --show-current 2>/dev/null || echo "unknown")

# Go version and module info
GO_VERSION = $(shell go version | cut -d ' ' -f 3)
MODULE_NAME = $(shell go list -m)

# Build flags
LDFLAGS = -s -w \
	-X '$(MODULE_NAME)/internal/version.Version=$(VERSION)' \
	-X '$(MODULE_NAME)/internal/version.BuildTime=$(BUILD_TIME)' \
	-X '$(MODULE_NAME)/internal/version.GitCommit=$(GIT_COMMIT)' \
	-X '$(MODULE_NAME)/internal/version.GitBranch=$(GIT_BRANCH)' \
	-X '$(MODULE_NAME)/internal/version.GoVersion=$(GO_VERSION)'

# Release build flags (more aggressive optimization)
RELEASE_LDFLAGS = $(LDFLAGS) -extldflags "-static"

# Directories
BUILD_DIR = bin
DIST_DIR = dist
COVERAGE_DIR = coverage

# Platforms for cross-compilation
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Colors for output
RED = \033[31m
GREEN = \033[32m
YELLOW = \033[33m
BLUE = \033[34m
NC = \033[0m # No Color

.PHONY: all build clean test coverage deps release release-all help

# Default target
all: clean deps test build

# Help target
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Dependencies
deps: ## Download and tidy dependencies
	@echo "$(BLUE)📦 Installing dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) tidy
	$(GOMOD) verify

# Build for development
build: ## Build the binary for current platform
	@echo "$(BLUE)🔨 Building $(BINARY_NAME)...$(NC)"
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v .

# Test
test: ## Run tests
	@echo "$(BLUE)🧪 Running tests...$(NC)"
	$(GOTEST) -v -race -timeout 30s ./...

# Test with coverage
coverage: ## Run tests with coverage
	@echo "$(BLUE)📊 Running tests with coverage...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_DIR)/coverage.html$(NC)"

# Lint (requires golangci-lint)
lint: ## Run linter
	@echo "$(BLUE)🔍 Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout 5m; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not found, skipping...$(NC)"; \
	fi

# Security scan (requires gosec)
security: ## Run security scan
	@echo "$(BLUE)🔒 Running security scan...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)⚠️  gosec not found, skipping...$(NC)"; \
	fi

# Pre-release checks
pre-release: clean deps lint security test coverage ## Run all checks before release
	@echo "$(GREEN)✅ All pre-release checks passed!$(NC)"

pre-release-skip-test: clean deps lint security
	@echo "$(GREEN)✅ All pre-release checks passed!$(NC)"

# Release build for current platform
release: pre-release-skip-test ## Build optimized release binary for current platform
	@echo "$(BLUE)🚀 Building release version $(VERSION)...$(NC)"
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 $(GOBUILD) \
		-a -installsuffix cgo \
		-ldflags "$(RELEASE_LDFLAGS)" \
		-o $(DIST_DIR)/$(BINARY_NAME) \
		.
	@echo "$(GREEN)✅ Release build completed: $(DIST_DIR)/$(BINARY_NAME)$(NC)"
	@echo "$(BLUE)📊 Binary info:$(NC)"
	@ls -lh $(DIST_DIR)/$(BINARY_NAME)
	@file $(DIST_DIR)/$(BINARY_NAME) 2>/dev/null || true

# Cross-platform release builds
release-all: pre-release ## Build release binaries for all platforms
	@echo "$(BLUE)🌍 Building release for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		output_name=$(BINARY_NAME)-$(VERSION)-$${GOOS}-$${GOARCH}; \
		if [ $$GOOS = "windows" ]; then output_name=$${output_name}.exe; fi; \
		echo "$(YELLOW)Building for $$GOOS/$$GOARCH...$(NC)"; \
		env CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH \
		$(GOBUILD) -a -installsuffix cgo \
		-ldflags "$(RELEASE_LDFLAGS)" \
		-o $(DIST_DIR)/$$output_name ./cmd/$(BINARY_NAME); \
		if [ $$? -ne 0 ]; then \
			echo "$(RED)❌ Failed to build for $$GOOS/$$GOARCH$(NC)"; \
			exit 1; \
		fi; \
	done
	@echo "$(GREEN)✅ All platform builds completed!$(NC)"
	@ls -lh $(DIST_DIR)/

# Create release archive
package: release-all ## Create release archives
	@echo "$(BLUE)📦 Creating release packages...$(NC)"
	@mkdir -p $(DIST_DIR)/packages
	@for binary in $(DIST_DIR)/$(BINARY_NAME)-*; do \
		if [ -f "$$binary" ]; then \
			basename=$$(basename $$binary); \
			if [[ $$basename == *".exe" ]]; then \
				zip -j $(DIST_DIR)/packages/$${basename%.*}.zip $$binary README.md LICENSE 2>/dev/null || true; \
			else \
				tar -czf $(DIST_DIR)/packages/$$basename.tar.gz -C $(DIST_DIR) $$(basename $$binary) -C .. README.md LICENSE 2>/dev/null || true; \
			fi; \
		fi; \
	done
	@echo "$(GREEN)✅ Release packages created in $(DIST_DIR)/packages/$(NC)"

# Docker build (if Dockerfile exists)
docker: ## Build Docker image
	@if [ -f "Dockerfile" ]; then \
		echo "$(BLUE)🐳 Building Docker image...$(NC)"; \
		docker build -t $(BINARY_NAME):$(VERSION) .; \
		docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest; \
		echo "$(GREEN)✅ Docker image built: $(BINARY_NAME):$(VERSION)$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  Dockerfile not found, skipping Docker build$(NC)"; \
	fi

# Install binary to GOPATH/bin
install: release ## Install binary to GOPATH/bin
	@echo "$(BLUE)📥 Installing $(BINARY_NAME)...$(NC)"
	@cp $(DIST_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "$(GREEN)✅ $(BINARY_NAME) installed to $(GOPATH)/bin/$(NC)"

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "$(BLUE)🧹 Cleaning...$(NC)"
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR) $(DIST_DIR) $(COVERAGE_DIR)
	@echo "$(GREEN)✅ Clean completed$(NC)"

# Development server (if applicable)
dev: ## Run in development mode with auto-reload
	@if command -v air >/dev/null 2>&1; then \
		echo "$(BLUE)🔄 Starting development server with air...$(NC)"; \
		air; \
	else \
		echo "$(YELLOW)⚠️  air not found, running normally...$(NC)"; \
		$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME) && $(BUILD_DIR)/$(BINARY_NAME); \
	fi

# Show version info
version: ## Show version information
	@echo "$(BLUE)📋 Version Information:$(NC)"
	@echo "  Version:    $(VERSION)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Git Branch: $(GIT_BRANCH)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Go Version: $(GO_VERSION)"
	@echo "  Module:     $(MODULE_NAME)"

# Release workflow: full pipeline
release-pipeline: clean deps lint security test coverage release-all package ## Complete release pipeline
	@echo "$(GREEN)🎉 Release pipeline completed successfully!$(NC)"
	@echo "$(BLUE)📦 Release artifacts:$(NC)"
	@ls -la $(DIST_DIR)/
	@echo "$(BLUE)📦 Release packages:$(NC)"
	@ls -la $(DIST_DIR)/packages/ 2>/dev/null || echo "No packages created"