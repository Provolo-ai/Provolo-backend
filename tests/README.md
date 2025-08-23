# Tests Directory

This directory contains all the tests for the Provolo Backend project, organized in a clean and maintainable structure.

## 📁 Directory Structure

```
tests/
├── README.md                 # This file
├── main_test.go             # Main test configuration and setup
├── handlers/                # Handler tests
│   ├── auth_test.go        # Authentication handler tests
│   ├── health_test.go      # Health check handler tests
│   ├── optimizer_test.go   # Profile optimizer handler tests
│   └── payment_test.go     # Payment webhook handler tests
├── middleware/              # Middleware tests
│   └── auth_test.go        # Authentication middleware tests
├── routes/                  # Route tests
│   └── routes_test.go      # Route setup and configuration tests
├── types/                   # Type tests
│   └── response_test.go    # Response type tests
├── integration/             # Integration tests (future)
└── utils/                   # Utility tests (future)
```

## 🧪 Test Organization

### Why This Structure?

1. **Separation of Concerns**: Tests are separated from source code, making the codebase cleaner
2. **Easy Navigation**: All tests are in one place, easy to find and maintain
3. **Scalability**: Easy to add new test categories without cluttering source code
4. **CI/CD Friendly**: Clear structure for automated testing pipelines
5. **Go Best Practices**: Follows Go testing conventions and patterns

### Test Categories

- **Handlers**: API endpoint logic testing
- **Middleware**: Authentication, CORS, and other middleware testing
- **Routes**: Route configuration and setup testing
- **Types**: Data structure and response type testing
- **Integration**: End-to-end API testing (planned)
- **Utils**: Utility function testing (planned)

## 🚀 Running Tests

### From Project Root

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test suites
make test-handlers
make test-middleware
make test-routes
make test-types

# Run tests with verbose output
make test-verbose
```

### From Tests Directory

```bash
# Run all tests in tests directory
go test ./...

# Run specific test package
go test ./handlers/
go test ./middleware/
go test ./routes/
go test ./types/

# Run with coverage
go test -coverprofile=coverage.out ./...
```

### Using Test Script

```bash
# Make script executable
chmod +x run_tests.sh

# Run comprehensive test suite
./run_tests.sh
```

## 📊 Test Coverage

### Current Coverage

- **Handlers**: ✅ Complete
  - Authentication (login, verify, logout)
  - Health check
  - Profile optimization
  - Payment webhook
- **Middleware**: ✅ Complete
  - Authentication middleware
  - CORS handling
- **Routes**: ✅ Complete
  - Route setup
  - Endpoint configuration
- **Types**: ✅ Complete
  - Response structures
  - JSON serialization

### Coverage Targets

- **Minimum**: 80%
- **Target**: 90%
- **Current**: Check with `make test-coverage`

## 🔧 Test Configuration

### Environment Variables

Tests automatically set these environment variables:

```bash
ENVIRONMENT=test
PORT=8080
FIREBASE_ENCODED_CONFIG=test-config
FIREBASE_SECRET_KEY=test-secret
GEMINI_API_KEY=test-gemini-key
MAX_PROMPT_LIMIT=2
```

### Test Dependencies

- `github.com/stretchr/testify/assert` - Assertions
- `github.com/stretchr/testify/mock` - Mocking
- `github.com/gin-gonic/gin` - HTTP testing

## 🎭 Mocking Strategy

### Firebase Authentication

```go
type MockAuthClient struct {
    mock.Mock
}

func (m *MockAuthClient) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
    args := m.Called(ctx, idToken)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*auth.Token), args.Error(1)
}
```

### Usage Example

```go
mockClient := &MockAuthClient{}
mockClient.On("VerifyIDToken", mock.Anything, "valid-token").Return(mockToken, nil)
```

## 📝 Adding New Tests

### 1. Create Test File

```bash
# Create new test file in appropriate directory
touch tests/handlers/new_handler_test.go
```

### 2. Follow Naming Convention

```go
package handlers

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestNewHandler_Functionality(t *testing.T) {
    // Test implementation
}
```

### 3. Test Structure

```go
func TestFunctionName_Scenario(t *testing.T) {
    // Setup
    // ... test setup code ...
    
    // Execute
    // ... test execution code ...
    
    // Assertions
    // ... test assertions ...
}
```

### 4. Add to Test Suite

Update the appropriate Makefile target if needed:

```makefile
test-handlers: deps
    @echo "Running handler tests..."
    go test -v ./tests/handlers/
```

## 🚨 Common Issues

### Import Paths

Make sure import paths are correct:

```go
// Correct
import "provolo-api/internal/handlers"

// Wrong
import "./handlers"
```

### Package Names

Test files should have the same package name as the source:

```go
// For testing internal/handlers/auth.go
package handlers

// For testing internal/middleware/auth.go
package middleware
```

### Mock Interfaces

Ensure mock objects implement the correct interfaces:

```go
type MockAuthClient struct {
    mock.Mock
}

// Implement all required methods
func (m *MockAuthClient) MethodName() error {
    args := m.Called()
    return args.Error(0)
}
```

## 🔍 Debugging Tests

### Verbose Output

```bash
go test -v ./tests/handlers/
```

### Run Specific Test

```bash
go test -v -run TestFunctionName ./tests/handlers/
```

### Test Logs

```bash
# Run tests and save output
go test -v ./tests/handlers/ 2>&1 | tee test-output.log
```

## 📈 Performance Testing

### Benchmarks

```bash
# Run all benchmarks
make bench

# Run specific package benchmarks
go test -bench=. ./tests/handlers/
```

### Example Benchmark

```go
func BenchmarkFunctionName(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Benchmark code here
    }
}
```

## 🚀 CI/CD Integration

### GitHub Actions

```yaml
- name: Run Tests
  run: |
    make test-ci
    make coverage-summary
```

### Test Commands for CI

```bash
# Quick test run for CI/CD
make test-ci

# Run tests with race detection
make test-race
```

## 📚 Additional Resources

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Testify Framework](https://github.com/stretchr/testify)
- [Go Testing Best Practices](https://golang.org/doc/code.html#Testing)
- [Project Testing Guide](../TESTING.md)

---

**Happy Testing! 🧪✨**
