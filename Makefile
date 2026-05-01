# Makefile for review-info cross-platform builds

# Variables
BINARY_NAME=review-info
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR=bin
GO=go
LDFLAGS=-ldflags="-s -w"

# Build targets
.PHONY: all build-all build-windows build-macos-amd64 build-macos-arm64 build-linux clean

all: build-all

build-all: build-windows build-macos-amd64 build-macos-arm64 build-linux

build-windows:
	@echo "Building for Windows amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/review-me/

build-macos-amd64:
	@echo "Building for macOS amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/review-me/

build-macos-arm64:
	@echo "Building for macOS arm64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/review-me/

build-linux:
	@echo "Building for Linux amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/review-me/

clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)


build-and-run: clean build-macos-arm64 
	./$(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64
