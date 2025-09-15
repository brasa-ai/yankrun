.PHONY: all build clean version

GOOS_ARCH := linux/amd64 linux/arm64 linux/386 linux/arm darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 windows/386
DIST_DIR := dist

# Version information - can be overridden by environment variable
ifeq ($(origin VERSION), environment)
  # VERSION is set from environment
else
  VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
endif

BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS := -ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

all: build

build:
	@echo "Building binaries..."
	@echo "Version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git commit: $(GIT_COMMIT)"
	@mkdir -p $(DIST_DIR)
	@for t in $(GOOS_ARCH); do \
		os=$${t%/*}; arch=$${t#*/}; \
		bin_name=yankrun-$${os}-$${arch}; \
		if [ "$$os" = "windows" ]; then bin_name="$${bin_name}.exe"; fi; \
		bin_path=$(DIST_DIR)/$$bin_name; \
		echo "  Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $$bin_path .; \
	done
	@echo "Build complete. Binaries in $(DIST_DIR)/"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(DIST_DIR)
	@echo "Clean complete."

version:
	@echo "Current version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git commit: $(GIT_COMMIT)"

# Create a git tag for the current version
tag:
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "Error: Cannot tag dev version. Please set VERSION environment variable."; \
		exit 1; \
	fi
	@echo "Creating git tag: $(VERSION)"
	git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "Tag created. Push with: git push origin $(VERSION)"

# Build and test before release
release-check: build
	@echo "Running tests..."
	go test ./...
	@echo "All tests passed. Ready for release $(VERSION)"


