G := $(shell go version | cut -d' ' -f 3,4 | sed 's/ /_/g')
V := $(shell git rev-parse --short HEAD)
APPVER := $(shell grep 'Version =' internal/version/version.go | cut -d \" -f2)
PWD := $(shell pwd)
GOPATH := $(shell go env GOPATH)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LD_FLAGS := -trimpath \
	-ldflags="-s -w \
	-X=github.com/sensepost/gowitness/internal/version.GitHash=$(V) \
	-X=github.com/sensepost/gowitness/internal/version.GoBuildEnv=$(G) \
	-X=github.com/sensepost/gowitness/internal/version.GoBuildTime=$(BUILD_TIME)"
BIN_DIR := build
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 linux/arm windows/amd64 windows/arm64
CGO := CGO_ENABLED=0

.PHONY: help prerequisites check-libpcap install-naabu build-with-deps

# Show available targets
help:
	@echo "Available targets:"
	@echo "  prerequisites    - Install libpcap and naabu portscanner"
	@echo "  check-libpcap   - Check if libpcap is installed"
	@echo "  install-naabu   - Install naabu portscanner"
	@echo "  build-with-deps - Full build including all dependencies"
	@echo "  clean           - Clean build artifacts"
	@echo "  test            - Run tests"
	@echo "  frontend        - Build frontend"
	@echo "  api-doc         - Generate API documentation"
	@echo "  build           - Build for all platforms"
	@echo "  integrity       - Generate checksums"

# Default target
default: clean test frontend api-doc build integrity

# Install prerequisites including naabu portscanner
prerequisites: check-libpcap install-naabu

# Check if libpcap is installed (required for naabu)
check-libpcap:
	@echo "Checking for libpcap..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		if ! brew list libpcap >/dev/null 2>&1; then \
			echo "libpcap not found. Installing via homebrew..."; \
			brew install libpcap; \
		else \
			echo "libpcap already installed"; \
		fi; \
	elif [ "$$(uname)" = "Linux" ]; then \
		if ! pkg-config --exists libpcap; then \
			echo "libpcap-dev not found. Please install with: sudo apt install -y libpcap-dev"; \
			exit 1; \
		else \
			echo "libpcap-dev already installed"; \
		fi; \
	else \
		echo "Unsupported platform for automatic libpcap installation"; \
		exit 1; \
	fi

# Install naabu portscanner
install-naabu:
	@echo "Installing naabu portscanner..."
	@if ! command -v naabu >/dev/null 2>&1; then \
		go install -v github.com/projectdiscovery/naabu/v2/cmd/naabu@latest; \
		echo "naabu installed successfully"; \
	else \
		echo "naabu already installed"; \
	fi

# Full build with prerequisites
build-with-deps: prerequisites clean test frontend api-doc build integrity

# Clean up build artifacts
clean:
	find $(BIN_DIR) -type f -name 'gowitness-*' -delete || true
	go clean -x

# Build frontend
frontend: check-npm
	@echo "Building frontend..."
	cd web/ui && npm i && npm run build

# Check if npm is installed
check-npm:
	@command -v npm >/dev/null 2>&1 || { echo >&2 "npm is not installed. Please install npm first."; exit 1; }

# Generate a swagger.json used for the api documentation
api-doc:
	go install github.com/swaggo/swag/cmd/swag@latest
	$(GOPATH)/bin/swag i --exclude ./web/ui --output web/docs
	$(GOPATH)/bin/swag f

# Run any tests
test:
	@echo "Running tests..."
	go test ./...

# Build for all platforms
build: $(PLATFORMS)

# Generic build target for platforms
$(PLATFORMS):
	$(eval GOOS=$(firstword $(subst /, ,$@)))
	$(eval GOARCH=$(lastword $(subst /, ,$@)))
	$(CGO) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LD_FLAGS) -o '$(BIN_DIR)/gowitness-$(APPVER)-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe)'

# Checksum integrity
integrity:
	cd $(BIN_DIR) && shasum *
