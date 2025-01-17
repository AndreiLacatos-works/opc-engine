# Define the name of your binary and other variables
BINARY_NAME=opc-engine-simulator
OUTPUT_DIR=build
RELEASE_DIR=release

# Release version
VERSION := $(shell cat versionfile)

# Development version
COMMIT_HASH := $(shell git rev-parse --short HEAD)
DEV_VERSION := development@$(COMMIT_HASH)

# Get the current timestamp
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')

# Default target executed when `make` is run without arguments
all: build

$(OUTPUT_DIR):
	mkdir -p $(OUTPUT_DIR)

$(RELEASE_DIR):
	mkdir -p $(RELEASE_DIR)

# Build the application
.PHONY: build
build: $(OUTPUT_DIR)
	@echo "Building the app..."
	go build -ldflags "\
		-X 'github.com/AndreiLacatos/opc-engine/config.version=$(DEV_VERSION)' \
		-X 'github.com/AndreiLacatos/opc-engine/config.buildTime=$(BUILD_TIME)'" \
		-o $(OUTPUT_DIR)/$(BINARY_NAME)
	@echo "Build completed."

# Clean up build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	go clean
	rm -rf $(OUTPUT_DIR)
	rm -rf $(RELEASE_DIR)
	@echo "Cleanup completed."

# Build & run the application
.PHONY: run
run: build
	@echo "Running the application..."
	./$(OUTPUT_DIR)/$(BINARY_NAME)

# Run all tests
.PHONY: test
test: 
	@echo "Running tests..."
	go test ./...

# Create new application release without version bump
.PHONY: build-release
build-release: $(RELEASE_DIR)
	@echo "Creating new release $(BINARY_NAME) version v$(VERSION)"
	go build -ldflags "\
		-X 'github.com/AndreiLacatos/opc-engine/config.version=$(VERSION)' \
		-X 'github.com/AndreiLacatos/opc-engine/config.buildTime=$(BUILD_TIME)' \
		-s -w" \
		-o $(RELEASE_DIR)/$(BINARY_NAME)
	@echo "Created new release"

# Create new application release with bumping patch version
.PHONY: build-release-bump-patch
build-release-bump-patch: $(RELEASE_DIR)
	@./versionbump.sh patch
	@make build-release

# Create new application release with bumping minor version
.PHONY: build-release-bump-minor
build-release-bump-minor: $(RELEASE_DIR)
	@./versionbump.sh minor
	@make build-release

# Create new application release with bumping major version
.PHONY: build-release-bump-major
build-release-bump-major: $(RELEASE_DIR)
	@./versionbump.sh major
	@make build-release
