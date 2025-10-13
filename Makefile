.PHONY: build run test clean install fmt lint help

# Binary name
BINARY_NAME=f6n
BINARY_PATH=./bin/$(BINARY_NAME)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build directory
BUILD_DIR=./bin

# Main package path
MAIN_PATH=./cmd/f6n

# Default target
all: build

## build: Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

## run: Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	script -q /dev/null $(BINARY_PATH)

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -f lambda-tui-debug.log
	@echo "Clean complete"

## install: Install the application to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(MAIN_PATH)
	@echo "Install complete"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "Format complete"

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: brew install golangci-lint" && exit 1)
	golangci-lint run ./...

## tidy: Tidy go modules
tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy
	@echo "Tidy complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Dependencies downloaded"

## upgrade: Upgrade dependencies
upgrade:
	@echo "Upgrading dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "Dependencies upgraded"

## dev: Run in development mode with dummy data
dev:
	@echo "Running in development mode..."
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_PATH) && $(BINARY_PATH)

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
