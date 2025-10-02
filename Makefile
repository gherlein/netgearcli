# Makefile for Netgear CLI Examples
# Builds example programs demonstrating the ntgrrc library

# Project information
PROJECT_NAME := netgearcli-examples
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOCLEAN := $(GOCMD) clean
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# Build flags
LDFLAGS := -ldflags "-X main.VERSION=$(VERSION) -X main.BUILD_TIME=$(BUILD_TIME) -X main.GIT_COMMIT=$(GIT_COMMIT)"
BUILD_FLAGS := $(LDFLAGS) -trimpath

# Directories
BUILD_DIR := bin
RELEASE_DIR := releases

# Example programs and their paths
EXAMPLES := poe-status poe-status-simple poe-management
CMD_DIRS := cmd/poe-status cmd/poe-status-simple cmd/poe-management

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
.PHONY: help
help:
	@echo "$(PROJECT_NAME) - Netgear CLI Examples"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build all example programs
.PHONY: build
build: deps
	@echo "Building netgearcli examples..."
	@mkdir -p $(BUILD_DIR)
	@for example in $(EXAMPLES); do \
		echo "Building $$example..."; \
		$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$$example ./cmd/$$example; \
		if [ $$? -ne 0 ]; then \
			echo "❌ Failed to build $$example"; \
			exit 1; \
		fi; \
	done
	@echo "✅ Build complete. Binaries in $(BUILD_DIR)/"

## clean: Remove build artifacts and temporary files
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	@echo "✅ Clean complete"

## deps: Download and verify dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download && $(GOMOD) verify
	@echo "✅ Dependencies ready"

## tidy: Clean up go.mod and go.sum
.PHONY: tidy
tidy:
	@echo "Tidying go module..."
	$(GOMOD) tidy
	@echo "✅ Module tidied"

## fmt: Format Go source code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "✅ Code formatted"

## vet: Run go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "✅ Vet complete"

## test: Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...
	@echo "✅ Tests complete"

## lint: Run basic linting (fmt + vet)
.PHONY: lint
lint: fmt vet
	@echo "✅ Basic linting complete"

## dev: Set up development environment
.PHONY: dev
dev: deps build
	@echo "✅ Development environment ready"
	@echo "Available binaries:"
	@ls -la $(BUILD_DIR)/

## check: Run all quality checks
.PHONY: check
check: lint test
	@echo "✅ All quality checks passed"

## version: Show version information
.PHONY: version
version:
	@echo "$(PROJECT_NAME) version information:"
	@echo "  Version: $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Go Version: $$($(GOCMD) version)"

## size: Show binary sizes
.PHONY: size
size: build
	@echo "Binary sizes:"
	@if [ -d $(BUILD_DIR) ]; then \
		ls -lh $(BUILD_DIR)/ | grep -v "^total" | awk '{print "  " $$9 ": " $$5}'; \
	fi

## release: Create release archive for current OS
.PHONY: release
release: build
	@echo "Creating release archive..."
	@mkdir -p $(RELEASE_DIR)
	@OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
	TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	ARCHIVE="netgearcli-$${TAG}-$${OS}.zip"; \
	ARCHIVE_PATH="$(RELEASE_DIR)/$${ARCHIVE}"; \
	echo "  OS: $${OS}"; \
	echo "  Tag: $${TAG}"; \
	echo "  Archive: $${ARCHIVE_PATH}"; \
	cd $(BUILD_DIR) && zip -r ../$${ARCHIVE_PATH} * > /dev/null && cd ..; \
	if [ -f "$${ARCHIVE_PATH}" ]; then \
		echo "✅ Release archive created: $${ARCHIVE_PATH}"; \
		ls -lh $${ARCHIVE_PATH} | awk '{print "  Size: " $$5}'; \
	else \
		echo "❌ Failed to create release archive"; \
		exit 1; \
	fi

# Create necessary directories
$(BUILD_DIR):
	@mkdir -p $@

$(RELEASE_DIR):
	@mkdir -p $@

# Build directory creation
.PHONY: directories
directories: $(BUILD_DIR) $(RELEASE_DIR)