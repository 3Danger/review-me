# Makefile for review-info cross-platform builds

# Variables
BINARY_NAME := review-info
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR   := bin
GO          := go
LDFLAGS     := -ldflags="-s -w"

# Targets
.PHONY: all build-all clean build-and-run test vet lint tidy

all: build-all

# Build for all platforms using a single loop
build-all: $(BUILD_DIR)
	@echo "Building version: $(VERSION)"
	@for plat in darwin/amd64/darwin-amd64 darwin/arm64/darwin-arm64 linux/amd64/linux-amd64 windows/amd64/windows-amd64.exe; do \
		os=$$(echo $$plat | cut -d/ -f1); \
		arch=$$(echo $$plat | cut -d/ -f2); \
		suffix=$$(echo $$plat | cut -d/ -f3); \
		echo "  $$os/$$arch -> $(BUILD_DIR)/$(BINARY_NAME)-$$suffix"; \
		GOOS=$$os GOARCH=$$arch $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$$suffix ./cmd/review-me/; \
	done
	@echo "Done."

$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)

test:
	@echo "Running tests..."
	$(GO) test ./... -v

vet:
	@echo "Running go vet..."
	$(GO) vet ./...

lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./... || true

tidy:
	@echo "Running go mod tidy..."
	$(GO) mod tidy

build-and-run: clean build-all
	./$(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64
