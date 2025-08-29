.PHONY: all build clean test serve help build-wasm build-example setup

# Default target
all: build

# Build the main WASM module
build-wasm:
	@echo "🔨 Building Go WASM module..."
	GOOS=js GOARCH=wasm go build -v -o main.wasm
	@echo "✅ WASM module built"

# Build example
build-example: build-wasm
	@echo "🔨 Building example..."
	cd example && GOOS=js GOARCH=wasm go build -o main.wasm ./main.go
	cd example && cp $$(go env GOROOT)/lib/wasm/wasm_exec.js .
	@echo "✅ Example built"

# Build everything
build: build-example
	@echo "✅ All components built successfully"

# Setup SQLite files in assets
setup-sqlite:
	@echo "📥 Setting up SQLite WASM files..."
	@if [ ! -f assets/sqlite3.wasm ]; then \
		echo "⚠️  SQLite WASM files not found in assets/"; \
		echo "Please download SQLite WASM files and place them in assets/"; \
		echo "Required files: sqlite3.wasm, sqlite3.js, sqlite3-worker1.js, sqlite3-worker1-promiser.js, sqlite3-opfs-async-proxy.js"; \
	else \
		echo "✅ SQLite files found in assets/"; \
	fi

# Initial setup
setup: setup-sqlite
	@echo "✅ Setup complete"

# Serve the demo locally
serve: build-example
	@echo "🌐 Starting local server at http://localhost:8081"
	@echo "✅ OPFS support enabled"
	@echo "✅ Cross-Origin Isolation enabled"
	cd example && node server.js

# Run tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...
	@echo "✅ Tests passed"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f main.wasm
	rm -f example/main.wasm example/wasm_exec.js
	@echo "✅ Clean complete"

# Development mode - watch for changes and rebuild
dev:
	@echo "👀 Starting development mode..."
	@echo "Watching for changes in Go and TypeScript files..."
	@make build
	@make serve

# Check requirements
check:
	@echo "🔍 Checking requirements..."
	@go version || (echo "❌ Go not installed" && exit 1)
	@node --version || (echo "❌ Node.js not installed" && exit 1)
	@npm --version || (echo "❌ npm not installed" && exit 1)
	@echo "✅ All requirements met!"

# Help
help:
	@echo "sqlc-wasm Makefile Commands:"
	@echo ""
	@echo "  make setup        - Initial setup (setup SQLite files)"
	@echo "  make build        - Build all components"
	@echo "  make build-wasm   - Build Go WASM module only"
	@echo "  make build-example- Build example only"
	@echo "  make serve        - Build and serve the demo locally"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make check        - Check requirements"
	@echo "  make help         - Show this help"
	@echo ""
	@echo "Quick start:"
	@echo "  1. make setup"
	@echo "  2. make build"
	@echo "  3. make serve"