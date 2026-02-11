export CGO_ENABLED=0

BINARY_NAME=webc
BUILD_DIR=bin

PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Build every platform's binary
.PHONY: build-all
build-all: $(PLATFORMS)

$(PLATFORMS):
	$(eval OS := $(word 1,$(subst /, ,$@)))
	$(eval ARCH := $(word 2,$(subst /, ,$@)))
	@echo "Building for $(OS)/$(ARCH)..."
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME)-$(OS)-$(ARCH)$(if $(filter windows,$(OS)),.exe) .

# Builds for a specific operating system/architecture
.PHONY: build-target
build-target:
	@if [ -z "$(OS)" ] || [ -z "$(ARCH)" ]; then \
		echo "Error: You must specify OS and ARCH. Example: make build-target OS=linux ARCH=amd64"; \
		exit 1; \
	fi
	$(eval EXTENSION := $(if $(filter windows,$(OS)),.exe,))
	@echo "Building for $(OS)/$(ARCH)..."
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME)-$(OS)-$(ARCH)$(EXTENSION) .

# Builds the binary for the current platform
.PHONY: build
build:
	@echo "Building for local host..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

# Runs the formatters
.PHONY: foramt
format:
	@go fmt ./...

# Runs the unit/integration tests
.PHONY: test
test:
	@go test -v ./...