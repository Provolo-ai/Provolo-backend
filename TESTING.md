# Testing Guide for Provolo Backend

This document provides comprehensive information about the testing setup for the Provolo Backend project.

## 🧪 Test Overview

The Provolo Backend includes a comprehensive test suite covering:

- **Handlers**: All API endpoint handlers
- **Middleware**: Authentication and CORS middleware
- **Routes**: Route configuration and setup
- **Types**: Response types and data structures
- **Integration**: End-to-end API testing

## 📋 Test Coverage

### Endpoints Tested

| Endpoint | Method | Auth Required | Test Coverage |
|----------|--------|---------------|---------------|
| `/api/v1/health` | GET | No | ✅ Complete |
| `/api/v1/auth/login` | POST | No | ✅ Complete |
| `/api/v1/auth/verify` | GET | No | ✅ Complete |
| `/api/v1/auth/logout` | POST | No | ✅ Complete |
| `/api/v1/protected/profile` | GET | Yes | ✅ Complete |
| `/api/v1/optimize-profile` | POST | Yes | ✅ Complete |
| `/api/v1/payment-webhook` | POST | No | ✅ Complete |

### Test Categories

- **Unit Tests**: Individual function testing
- **Integration Tests**: API endpoint testing
- **Middleware Tests**: Authentication and CORS testing
- **Validation Tests**: Input validation and sanitization
- **Error Handling Tests**: Error scenarios and edge cases

## 🚀 Quick Start

### Prerequisites

- Go 1.24.4 or higher
- Git

### Install Dependencies

```bash
# Install Go dependencies
go mod tidy

# Install test dependencies
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock
```

### Run All Tests

```bash
# Using Make
make test

# Using Go directly
go test ./...

# Using the test script
chmod +x run_tests.sh
./run_tests.sh
```

## 🛠️ Test Commands

### Using Make (Recommended)

```bash
# Show all available commands
make help

# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run tests with coverage report
make test-coverage

# Run specific test suites
make test-handlers
make test-middleware
make test-routes
make test-types

# Run benchmarks
make bench

# Clean test artifacts
make clean
```

### Using Go Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -coverprofile=coverage.out -covermode=atomic ./...

# Run specific package tests
go test -v ./tests/handlers/
go test -v ./tests/middleware/
go test -v ./tests/routes/
go test -v ./tests/types/

# Run tests with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

## 📊 Coverage Reports

### Generate Coverage Report

```bash
# Generate coverage data
go test -coverprofile=coverage.out -covermode=atomic ./...

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Show coverage summary
go tool cover -func=coverage.out
```

### Coverage Targets

- **Minimum Coverage**: 80%
- **Target Coverage**: 90%
- **Current Coverage**: Check with `make test-coverage`

## 🧩 Test Structure

### Test Files

```
tests/
├── README.md                 # Tests directory documentation
├── main_test.go             # Main test configuration and setup
├── handlers/                # Handler tests
│   ├── auth_test.go        # Authentication handler tests
│   ├── health_test.go      # Health check tests
│   ├── optimizer_test.go   # Profile optimizer tests
│   └── payment_test.go     # Payment webhook tests
├── middleware/              # Middleware tests
│   └── auth_test.go        # Auth middleware tests
├── routes/                  # Route tests
│   └── routes_test.go      # Route setup tests
├── types/                   # Type tests
│   └── response_test.go    # Response type tests
├── integration/             # Integration tests (future)
└── utils/                   # Utility tests (future)
```

### Test Naming Convention

- **Test Functions**: `Test[FunctionName]_[Scenario]`
- **Test Files**: `[package]_test.go`
- **Mock Types**: `Mock[TypeName]`

### Example Test Structure

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

## 🔧 Test Configuration

### Environment Variables

Tests use the following environment variables:

```bash
ENVIRONMENT=test
PORT=8080
FIREBASE_ENCODED_CONFIG=test-config
FIREBASE_SECRET_KEY=test-secret
GEMINI_API_KEY=test-gemini-key
MAX_PROMPT_LIMIT=2
```

### Test Configuration File

The `test_config.go` file automatically sets up the test environment.

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

### Usage in Tests

```go
mockClient := &MockAuthClient{}
mockClient.On("VerifyIDToken", mock.Anything, "valid-token").Return(mockToken, nil)
```

## 🧪 Test Scenarios

### Authentication Tests

- ✅ Valid login with Firebase ID token
- ✅ Invalid token rejection
- ✅ Session verification
- ✅ Logout functionality
- ✅ Missing authentication handling

### Input Validation Tests

- ✅ Empty field validation
- ✅ Length limit validation
- ✅ Suspicious content detection
- ✅ Whitespace handling

### Error Handling Tests

- ✅ Invalid JSON payloads
- ✅ Missing required fields
- ✅ Rate limiting
- ✅ Service failures

### CORS Tests

- ✅ Production CORS restrictions
- ✅ Development CORS flexibility
- ✅ Header validation

## 🚨 Common Test Issues

### Firebase Dependencies

Some tests are skipped due to Firebase dependency complexity:

```go
t.Skip("Skipping ProfileOptimizer test as it requires complex mocking of Firebase and Gemini services")
```

### Solution: Use Dependency Injection

Consider refactoring to use interfaces for better testability:

```go
type AuthService interface {
    VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
    VerifySessionCookie(ctx context.Context, sessionCookie string) (*auth.Token, error)
}
```

## 📈 Performance Testing

### Benchmarks

```bash
# Run all benchmarks
make bench

# Run specific package benchmarks
go test -bench=. ./internal/handlers/
```

### Example Benchmark

```go
func BenchmarkProfileOptimizer(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Benchmark code here
    }
}
```

## 🔍 Debugging Tests

### Verbose Output

```bash
go test -v ./...
```

### Test Logs

```bash
# Run tests and save output
go test -v ./... 2>&1 | tee test-output.log
```

### Individual Test Debugging

```bash
# Run specific test
go test -v -run TestFunctionName ./...
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

### Go Testing Documentation

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Go Test Coverage](https://blog.golang.org/cover)
- [Go Testing Best Practices](https://golang.org/doc/code.html#Testing)

### Testify Framework

- [Testify Documentation](https://github.com/stretchr/testify)
- [Mock Examples](https://github.com/stretchr/testify#mock-package)

### Firebase Testing

- [Firebase Admin SDK Testing](https://firebase.google.com/docs/admin/setup#initialize-sdk)
- [Firebase Emulator Suite](https://firebase.google.com/docs/emulator-suite)

## 🤝 Contributing to Tests

### Adding New Tests

1. Create test file: `[package]_test.go`
2. Follow naming convention: `Test[FunctionName]_[Scenario]`
3. Include setup, execution, and assertions
4. Add to appropriate test suite

### Test Quality Checklist

- [ ] Test covers happy path
- [ ] Test covers error cases
- [ ] Test uses appropriate mocks
- [ ] Test has clear assertions
- [ ] Test follows naming convention
- [ ] Test includes comments for complex logic

### Running Tests Before Commit

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Check for any failures
make test-verbose
```

## 📞 Support

For testing-related questions or issues:

1. Check this documentation
2. Review existing test examples
3. Check test output logs
4. Create an issue with test details

---

**Happy Testing! 🧪✨**
