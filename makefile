# Name of the binary executable
BINARY_NAME := test

# Go source files
SRC := $(shell find . -type f -name '*.go')

# Default target to build the project
all: build

# Build the project
build: $(SRC)
	@echo "Building the project..."
	go build -o $(BINARY_NAME)

# Run the project
run: build
	@echo "Running the project..."
	./$(BINARY_NAME)

# Test the project
test:
	@echo "Running tests..."
	go test ./...

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	go clean
	rm -f $(BINARY_NAME)

# Install the binary to GOPATH/bin
install: build
	@echo "Installing..."
	go install

# Format the source code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint the source code
lint:
	@echo "Linting code..."
	golangci-lint run

# Run a complete cycle of build, test, and run
.PHONY: all build run test clean install fmt lint
