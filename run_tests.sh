#!/bin/bash

# Provolo Backend Test Runner
# This script runs all tests with coverage and generates reports

echo "🚀 Starting Provolo Backend Test Suite..."
echo "=========================================="

# Set test environment
export ENVIRONMENT=test
export PORT=8080

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go first."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
print_status "Go version: $GO_VERSION"

# Clean previous test artifacts
print_status "Cleaning previous test artifacts..."
rm -rf coverage.out coverage.html
rm -rf test-results/

# Create test results directory
mkdir -p test-results

# Install test dependencies
print_status "Installing test dependencies..."
go mod tidy
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock

# Run tests with verbose output and coverage
print_status "Running tests with coverage..."
go test -v -coverprofile=coverage.out -covermode=atomic ./... 2>&1 | tee test-results/test-output.log

# Check if tests passed
TEST_EXIT_CODE=${PIPESTATUS[0]}
if [ $TEST_EXIT_CODE -eq 0 ]; then
    print_success "All tests passed!"
else
    print_error "Some tests failed. Check test-results/test-output.log for details."
fi

# Generate coverage report
if [ -f coverage.out ]; then
    print_status "Generating coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    
    # Show coverage summary
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    print_status "Code coverage: ${COVERAGE}%"
    
    if (( $(echo "$COVERAGE >= 80" | bc -l) )); then
        print_success "Coverage is above 80% threshold"
    else
        print_warning "Coverage is below 80% threshold"
    fi
else
    print_warning "No coverage data generated"
fi

# Run specific test suites
echo ""
echo "📋 Running Individual Test Suites..."
echo "===================================="

# Test handlers
print_status "Testing handlers..."
go test -v ./tests/handlers/ 2>&1 | tee test-results/handlers-test.log

# Test middleware
print_status "Testing middleware..."
go test -v ./tests/middleware/ 2>&1 | tee test-results/middleware-test.log

# Test routes
print_status "Testing routes..."
go test -v ./tests/routes/ 2>&1 | tee test-results/routes-test.log

# Test types
print_status "Testing types..."
go test -v ./tests/types/ 2>&1 | tee test-results/types-test.log

# Run benchmarks if available
echo ""
echo "🏃 Running Benchmarks..."
echo "========================"
go test -bench=. ./... 2>&1 | tee test-results/benchmarks.log

# Generate test summary
echo ""
echo "📊 Test Summary"
echo "==============="
echo "Test output: test-results/test-output.log"
echo "Coverage report: coverage.html"
echo "Individual test logs: test-results/"

# Check for any test failures in logs
FAILED_TESTS=$(grep -r "FAIL" test-results/ | wc -l)
if [ $FAILED_TESTS -gt 0 ]; then
    print_error "Found $FAILED_TESTS test failures. Check the logs above."
    exit 1
else
    print_success "No test failures found!"
fi

echo ""
print_success "Test suite completed successfully!"
echo "Check test-results/ directory for detailed logs and coverage.html for coverage report."
