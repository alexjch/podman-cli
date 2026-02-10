.PHONY: all build test clean install fmt vet coverage help

# Binary name
BINARY_NAME=podman-cli
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build parameters
MAIN_PATH=./cmd/$(BINARY_NAME)
MAIN_FILE=$(MAIN_PATH)/main.go

# Default target
all: test build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -cover ./...

## coverage: Generate coverage report
coverage:
	@echo "Generating coverage report..."
	@mkdir -p coverage
	$(GOTEST) -coverprofile=coverage/coverage.out ./...
	$(GOCMD) tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report generated: coverage/coverage.html"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf coverage
	@echo "Clean complete"

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

## fmt: Format all Go source files
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

## mod-tidy: Tidy up go.mod and go.sum
mod-tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy

## mod-download: Download dependencies
mod-download:
	@echo "Downloading dependencies..."
	$(GOMOD) download

## run: Build and run the binary with default args
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

## check: Run fmt, vet, and test
check: fmt vet test

## help: Display this help message
help:
	@echo "Makefile for $(BINARY_NAME)"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
