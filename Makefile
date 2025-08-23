# Provolo Backend Makefile
# Common commands for development and testing

.PHONY: help test test-verbose test-coverage test-handlers test-middleware test-routes test-types clean build run lint

# Default target
help:
	@echo "Provolo Backend - Available Commands:"
	@echo ""
	@echo "Testing:"
	@echo "  test              - Run all tests"
	@echo "  test-verbose      - Run tests with verbose output"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  test-handlers     - Run only handler tests"
	@echo "  test-middleware   - Run only middleware tests"
	@echo "  test-routes       - Run only route tests"
	@echo "  test-types        - Run only type tests"
	@echo ""
	@echo "Development:"
	@echo "  build             - Build the application"
	@echo "  run               - Run the application"
	@echo "  clean             - Clean build artifacts and test files"
	@echo "  lint              - Run linting checks"
	@echo "  deps              - Install dependencies"
	@echo ""
	@echo "Utilities:"
	@echo "  help              - Show this help message"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go get github.com/stretchr/testify/assert
	go get github.com/stretchr/testify/mock

# Run all tests
test: deps
	@echo "Running all tests..."
	go test ./...

# Run tests with verbose output
test-verbose: deps
	@echo "Running tests with verbose output..."
	go test -v ./...

# Run tests with coverage
test-coverage: deps
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run specific test suites
test-handlers: deps
	@echo "Running handler tests..."
	go test -v ./tests/handlers/

test-middleware: deps
	@echo "Running middleware tests..."
	go test -v ./tests/middleware/

test-routes: deps
	@echo "Running route tests..."
	go test -v ./tests/routes/

test-types: deps
	@echo "Running type tests..."
	go test -v ./tests/types/

# Run benchmarks
bench: deps
	@echo "Running benchmarks..."
	go test -bench=. ./...

# Build the application
build:
	@echo "Building application..."
	go build -o bin/provolo-backend ./cmd/api

# Run the application
run: build
	@echo "Running application..."
	./bin/provolo-backend

# Clean build artifacts and test files
clean:
	@echo "Cleaning build artifacts and test files..."
	rm -rf bin/
	rm -rf coverage.out
	rm -rf coverage.html
	rm -rf test-results/
	go clean -testcache

# Run linting checks
lint:
	@echo "Running linting checks..."
	golangci-lint run

# Install linting tools
install-lint:
	@echo "Installing linting tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run tests with race detection
test-race: deps
	@echo "Running tests with race detection..."
	go test -race ./...

# Run tests with short timeout
test-short: deps
	@echo "Running tests with short timeout..."
	go test -short ./...

# Generate test coverage badge
coverage-badge: test-coverage
	@echo "Generating coverage badge..."
	@if command -v gocover-cobertura >/dev/null 2>&1; then \
		gocover-cobertura < coverage.out > coverage.xml; \
		echo "Coverage XML generated: coverage.xml"; \
	else \
		echo "gocover-cobertura not installed. Install with: go install github.com/t-yuki/gocover-cobertura@latest"; \
	fi

# Run integration tests
test-integration: deps
	@echo "Running integration tests..."
	go test -tags=integration ./...

# Show test coverage summary
coverage-summary: test-coverage
	@echo "Coverage summary:"
	go tool cover -func=coverage.out

# Run tests and show coverage
test-full: test-coverage coverage-summary
	@echo "Full test suite completed with coverage report"

# Development mode - run tests on file changes
dev-test:
	@echo "Starting development test mode (run tests on file changes)..."
	@if command -v air >/dev/null 2>&1; then \
		air -c .air.toml; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Or run: make install-air"; \
	fi

# Install Air for hot reloading
install-air:
	@echo "Installing Air for hot reloading..."
	go install github.com/cosmtrek/air@latest

# Run tests with specific environment
test-env: deps
	@echo "Running tests with test environment..."
	ENVIRONMENT=test PORT=8080 go test -v ./...

# Show test dependencies
test-deps:
	@echo "Test dependencies:"
	@go list -f '{{.Dir}}' ./... | xargs -I {} sh -c 'echo "{}:" && go list -f "{{.Imports}}" {} | grep -E "(test|mock|assert)" | head -5'

# Quick test run for CI/CD
test-ci: deps
	@echo "Running tests for CI/CD..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Tests completed for CI/CD"
