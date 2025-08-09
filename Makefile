.PHONY: all build clean

GOOS_ARCH := linux/amd64 linux/arm64 linux/386 linux/arm darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 windows/386
DIST_DIR := dist

all: build

build:
	@echo "Building binaries..."
	@mkdir -p $(DIST_DIR)
	@for t in $(GOOS_ARCH); do \
		os=$${t%/*}; arch=$${t#*/}; \
		bin_name=yankrun-$${os}-$${arch}; \
		if [ "$$os" = "windows" ]; then bin_name="$${bin_name}.exe"; fi; \
		bin_path=$(DIST_DIR)/$$bin_name; \
		echo "  Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build -ldflags="-s -w" -o $$bin_path .; \
	done
	@echo "Build complete. Binaries in $(DIST_DIR)/"

clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(DIST_DIR)
	@echo "Clean complete."


