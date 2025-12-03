.PHONY: build build-all clean test

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# Default target
build:
	go build $(LDFLAGS) -o bin/maestro-ios-device ./cmd/maestro-ios-device

# Build for all platforms
build-all: build-darwin-amd64 build-darwin-arm64

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/maestro-ios-device-darwin-amd64 ./cmd/maestro-ios-device

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/maestro-ios-device-darwin-arm64 ./cmd/maestro-ios-device

# Clean build artifacts
clean:
	rm -rf bin/ dist/

# Run tests
test:
	go test -v ./...

# Install locally
install: build
	cp bin/maestro-ios-device /usr/local/bin/

# Tidy dependencies
tidy:
	go mod tidy

# Download dependencies
deps:
	go mod download

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build for current platform"
	@echo "  build-all       - Build for all platforms (darwin/amd64, darwin/arm64)"
	@echo "  clean           - Remove build artifacts"
	@echo "  test            - Run tests"
	@echo "  install         - Install to /usr/local/bin"
	@echo "  tidy            - Tidy go.mod"
	@echo "  deps            - Download dependencies"
	@echo "  fmt             - Format code"
	@echo "  lint            - Run linter"
