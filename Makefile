.PHONY: help build run test test-watch test-coverage test-coverage-html clean fmt lint swagger check deps migrate-up bench audit update-deps

# Variables
APP_NAME=stock-api
BINARY_DIR=bin
MAIN_PATH=cmd/api/main.go
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Default target
help:
	@echo "Available targets:"
	@echo "  build              - Build the application binary"
	@echo "  run                - Run the application"
	@echo "  test               - Run all tests"
	@echo "  test-watch         - Run tests in watch mode"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  test-coverage-html - Generate HTML coverage report and open in browser"
	@echo "  clean              - Remove build artifacts"
	@echo "  fmt                - Format Go code"
	@echo "  lint               - Run golangci-lint"
	@echo "  swagger            - Generate Swagger documentation"
	@echo "  check              - Run all checks (fmt, lint, test)"
	@echo "  deps               - Install dependencies"
	@echo "  bench              - Run benchmark tests"
	@echo "  audit              - Run security audit"
	@echo "  update-deps        - Update dependencies"

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BINARY_DIR)
	@go build -o $(BINARY_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_DIR)/$(APP_NAME)"

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	@go run $(MAIN_PATH)

# Run all tests
test:
	@echo "Running tests..."
	@gotestsum --format testname -- -race ./...

# Run tests in watch mode (re-run on changes)
test-watch:
	@echo "Running tests in watch mode..."
	@gotestsum --watch --format testname -- -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@gotestsum --format testname -- -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@go tool cover -func=$(COVERAGE_FILE)
	@echo "Coverage report saved to $(COVERAGE_FILE)"

# Generate HTML coverage report
test-coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Opening coverage report in browser..."
	@start $(COVERAGE_HTML) 2>/dev/null || open $(COVERAGE_HTML) 2>/dev/null || xdg-open $(COVERAGE_HTML) 2>/dev/null

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BINARY_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@go clean
	@echo "Clean complete"

# Format Go code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/api/main.go -o docs || echo "Install swag: go install github.com/swaggo/swag/cmd/swag@latest"
	@echo "Swagger documentation generated in docs/"

# Run all checks
check: fmt lint test
	@echo "All checks passed!"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Database migrations (placeholder - implement as needed)
migrate-up:
	@echo "Running database migrations..."
	@echo "Migrations are automatically applied on startup"

# Benchmark tests
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Security audit
audit:
	@echo "Running security audit..."
	@go list -json -m all | nancy sleuth || echo "Install nancy: go install github.com/sonatype-nexus-community/nancy@latest"

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "Dependencies updated"
