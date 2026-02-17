.PHONY: build build-bin build-bin-linux run test clean install fmt lint help tidy deps dev-image

DOCKER_COMPOSE ?= docker compose
APP_SERVICE ?= app
DEV_SERVICE ?= dev
HOST_GOOS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
HOST_GOARCH := $(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/;s/arm64/arm64/')

# Default target
all: build

## build: Build the application
build:
	@echo "Building Docker image..."
	$(DOCKER_COMPOSE) build $(APP_SERVICE)
	@echo "Image ready: f6n:local"

## build-bin: Build host-architecture binary inside docker (no local Go)
build-bin:
	@echo "Building host binary (GOOS=$(HOST_GOOS) GOARCH=$(HOST_GOARCH)) inside container..."
	@mkdir -p ./bin
	$(DOCKER_COMPOSE) run --rm -e GOOS=$(HOST_GOOS) -e GOARCH=$(HOST_GOARCH) $(DEV_SERVICE) "cd /workspace && CGO_ENABLED=0 go build -o bin/f6n ./cmd/f6n"
	@echo "Binary written to ./bin/f6n"

## build-bin-linux: Build linux/amd64 binary by copying from built image
build-bin-linux: build
	@echo "Extracting linux/amd64 binary from image..."
	@mkdir -p ./bin
	@cid=$$($(DOCKER_COMPOSE) create $(APP_SERVICE)); \
	docker cp $$cid:/usr/local/bin/f6n ./bin/f6n; \
	$(DOCKER_COMPOSE) rm -f $$cid >/dev/null; \
	echo "Binary copied to ./bin/f6n (linux/amd64)"

## run: Run the application
run: build
	@echo "Launching TUI via docker compose (Ctrl+C to quit)..."
	$(DOCKER_COMPOSE) run --rm --service-ports $(APP_SERVICE)

## test: Run tests
test:
	@echo "Running tests in container..."
	$(DOCKER_COMPOSE) run --rm $(DEV_SERVICE) "cd /workspace && go test -v ./..."

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage in container..."
	$(DOCKER_COMPOSE) run --rm $(DEV_SERVICE) "cd /workspace && go test -v -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html"
	@echo "Coverage report generated: coverage.html"

## clean: Clean build artifacts
clean:
	@echo "Cleaning workspace..."
	rm -rf ./bin
	rm -f coverage.out coverage.html
	rm -f f6n-debug.log
	$(DOCKER_COMPOSE) down --volumes --remove-orphans
	@echo "Clean complete"

## install: Install the application to $GOPATH/bin
install: build-bin
	@echo "Binary available at ./bin/f6n"

## fmt: Format code
fmt:
	@echo "Running gofmt inside container..."
	$(DOCKER_COMPOSE) run --rm $(DEV_SERVICE) "cd /workspace && gofmt -w ."
	@echo "Format complete"

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running golangci-lint inside container..."
	$(DOCKER_COMPOSE) run --rm $(DEV_SERVICE) "cd /workspace && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && golangci-lint run ./..."

## tidy: Tidy go modules
tidy:
	@echo "Tidying modules inside container..."
	$(DOCKER_COMPOSE) run --rm $(DEV_SERVICE) "cd /workspace && go mod tidy"
	@echo "Tidy complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies inside container..."
	$(DOCKER_COMPOSE) run --rm $(DEV_SERVICE) "cd /workspace && go mod download"
	@echo "Dependencies downloaded"

## upgrade: Upgrade dependencies
upgrade:
	@echo "Upgrading dependencies inside container..."
	$(DOCKER_COMPOSE) run --rm $(DEV_SERVICE) "cd /workspace && go get -u ./... && go mod tidy"
	@echo "Dependencies upgraded"

## dev: Run in development mode with dummy data
dev: build
	@echo "Starting app container in interactive mode..."
	$(DOCKER_COMPOSE) run --rm --service-ports $(APP_SERVICE)

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
