# Makefile para o projeto dck
BINARY_NAME=dck
VERSION?=dev
BUILD_DIR=bin
CMD_DIR=cmd
PKG_NAME=github.com/yourusername/dck

# Build flags
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S') -s -w"
GOFLAGS=-trimpath

# Go related variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Detecta o OS para builds específicos
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    DETECTED_OS := linux
    INSTALL_PATH := /usr/local/bin
    USER_INSTALL_PATH := $(HOME)/bin
endif
ifeq ($(UNAME_S),Darwin)
    DETECTED_OS := darwin
    INSTALL_PATH := /usr/local/bin
    USER_INSTALL_PATH := $(HOME)/bin
endif
ifeq ($(OS),Windows_NT)
    DETECTED_OS := windows
    INSTALL_PATH := C:\Program Files\$(BINARY_NAME)
    USER_INSTALL_PATH := $(USERPROFILE)\bin
    BINARY_EXTENSION := .exe
else
    BINARY_EXTENSION :=
endif

.PHONY: all build install test clean dev deps tidy vendor generate help

# Target padrão
all: clean deps build

# Build principal
build:
	@echo "🔨 Building $(BINARY_NAME) for $(DETECTED_OS)..."
ifeq ($(DETECTED_OS),windows)
	@if not exist "$(BUILD_DIR)" mkdir $(BUILD_DIR)
	$(GOBUILD) $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXTENSION) main.go
else
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXTENSION) main.go
endif
	@echo "✅ Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXTENSION)"

# Build para múltiplas plataformas
build-all: clean
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe main.go
	@echo "✅ Multi-platform build completed"

# Instalação local (sistema)
install: build
	@echo "📦 Installing $(BINARY_NAME) for $(DETECTED_OS)..."
ifeq ($(DETECTED_OS),windows)
	@echo "Creating installation directory..."
	@if not exist "$(INSTALL_PATH)" mkdir "$(INSTALL_PATH)"
	copy "$(BUILD_DIR)\$(BINARY_NAME)$(BINARY_EXTENSION)" "$(INSTALL_PATH)\"
	@echo "✅ $(BINARY_NAME) installed to $(INSTALL_PATH)"
	@echo "💡 Add $(INSTALL_PATH) to your PATH environment variable"
	@echo "💡 Or run: setx PATH "$(INSTALL_PATH);%PATH%""
else
	sudo cp $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXTENSION) $(INSTALL_PATH)/
	@echo "✅ $(BINARY_NAME) installed to $(INSTALL_PATH)/"
endif

# Instalação sem privilégios administrativos (usuário)
install-user: build
	@echo "📦 Installing $(BINARY_NAME) for current user ($(DETECTED_OS))..."
ifeq ($(DETECTED_OS),windows)
	@if not exist "$(USER_INSTALL_PATH)" mkdir "$(USER_INSTALL_PATH)"
	copy "$(BUILD_DIR)\$(BINARY_NAME)$(BINARY_EXTENSION)" "$(USER_INSTALL_PATH)\"
	@echo "✅ $(BINARY_NAME) installed to $(USER_INSTALL_PATH)"
	@echo "💡 Add $(USER_INSTALL_PATH) to your PATH:"
	@echo "   setx PATH "$(USER_INSTALL_PATH);%PATH%""
else
	@mkdir -p $(USER_INSTALL_PATH)
	cp $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXTENSION) $(USER_INSTALL_PATH)/
	@echo "✅ $(BINARY_NAME) installed to $(USER_INSTALL_PATH)/"
	@echo "💡 Make sure $(USER_INSTALL_PATH) is in your PATH"
	@echo "💡 Add to ~/.bashrc or ~/.zshrc: export PATH=\"$(USER_INSTALL_PATH):\$PATH\""
endif

# Instalação automática baseada no OS
install-auto: build
ifeq ($(DETECTED_OS),windows)
	@echo "🔍 Windows detected - installing for current user..."
	@$(MAKE) install-user
else
	@echo "🔍 Unix-like system detected - attempting system installation..."
	@if command -v sudo >/dev/null 2>&1; then \
		$(MAKE) install; \
	else \
		echo "⚠️  sudo not available, installing for current user..."; \
		$(MAKE) install-user; \
	fi
endif

# Testes
test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "✅ Tests completed"

# Testes com coverage
test-coverage: test
	@echo "📊 Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Benchmark
bench:
	@echo "⚡ Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Limpeza
clean:
	@echo "🧹 Cleaning..."
	$(GOCLEAN)
ifeq ($(DETECTED_OS),windows)
	@if exist "$(BUILD_DIR)" rmdir /s /q $(BUILD_DIR)
	@if exist "coverage.out" del coverage.out
	@if exist "coverage.html" del coverage.html
else
	rm -rf $(BUILD_DIR)/
	rm -f coverage.out coverage.html
endif
	@echo "✅ Clean completed"

# Desenvolvimento (build e executa)
dev: build
	@echo "🚀 Running in development mode..."
ifeq ($(DETECTED_OS),windows)
	$(BUILD_DIR)\$(BINARY_NAME)$(BINARY_EXTENSION)
else
	./$(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXTENSION)
endif

# Gerenciamento de dependências
deps:
	@echo "📚 Downloading dependencies..."
	$(GOGET) -d ./...
	@echo "✅ Dependencies downloaded"

# Atualiza go.mod
tidy:
	@echo "🧹 Tidying go.mod..."
	$(GOMOD) tidy
	@echo "✅ go.mod tidied"

# Cria vendor directory
vendor:
	@echo "📦 Creating vendor directory..."
	$(GOMOD) vendor
	@echo "✅ Vendor directory created"

# Gera código (se necessário)
generate:
	@echo "⚙️  Generating code..."
	$(GOCMD) generate ./...
	@echo "✅ Code generation completed"

# Gera um novo comando usando o script
new-cmd:
	@if [ -z "$(CMD)" ]; then \
		echo "❌ Use: make new-cmd CMD=nome-do-comando"; \
		exit 1; \
	fi
	@echo "📝 Generating new command: $(CMD)"
	$(GOCMD) run tools/gen-command.go $(CMD)

# Formata o código
fmt:
	@echo "🎨 Formatting code..."
	$(GOCMD) fmt ./...
	@echo "✅ Code formatted"

# Linting (requer golangci-lint)
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Install it with:"; \
		echo "   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2"; \
	fi

# Verifica vulnerabilidades (requer govulncheck)
security:
	@echo "🔒 Checking for vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "⚠️  govulncheck not installed. Install it with:"; \
		echo "   go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# Workflow completo para CI/CD
ci: clean deps fmt lint test build
	@echo "✅ CI pipeline completed successfully"

# Release (example)
release: clean build-all test
	@echo "🚀 Creating release..."
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "dev" ]; then \
		echo "❌ Please set VERSION. Example: make release VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "✅ Release $(VERSION) ready in $(BUILD_DIR)/"

# Ajuda
help:
	@echo "🔧 Available commands:"
	@echo ""
	@echo "Build commands:"
	@echo "  build      - Build the binary"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  clean      - Clean build artifacts"
	@echo ""
	@echo "Development:"
	@echo "  dev        - Build and run in development mode"
	@echo "  new-cmd    - Generate new command (use CMD=name)"
	@echo "  fmt        - Format code"
	@echo "  generate   - Run go generate"
	@echo ""
	@echo "Testing:"
	@echo "  test       - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  bench      - Run benchmarks"
	@echo "  lint       - Run linter"
	@echo "  security   - Check vulnerabilities"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps       - Download dependencies"
	@echo "  tidy       - Tidy go.mod"
	@echo "  vendor     - Create vendor directory"
	@echo ""
	@echo "Installation:"
	@echo "  install      - Install binary to /usr/local/bin (requires sudo)"
	@echo "  install-user - Install binary to ~/bin"
	@echo ""
	@echo "CI/CD:"
	@echo "  ci         - Run complete CI pipeline"
	@echo "  release    - Create release (set VERSION=vX.X.X)"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make new-cmd CMD=user-create"
	@echo "  make release VERSION=v1.0.0"
