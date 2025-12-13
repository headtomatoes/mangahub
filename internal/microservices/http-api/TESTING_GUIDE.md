# Unit Test Suite for MangaHub HTTP API

## Test Files Created

### Handler Tests
1. **auth_handler_test.go** - Authentication handler tests (Register, Login, Refresh, Revoke)
2. **comment_handler_test.go** - Comment handler tests (Create, Update, Delete, Get, List)
3. **rating_handler_test.go** - Rating handler tests (Create/Update, Get, Delete, Average)
4. **library_handler_test.go** - Library handler tests (Add, Remove, List)
5. **mangaHandler_test.go** - Manga handler tests (CRUD operations, Search)

### Service Tests
1. **auth_service_test.go** - Authentication service tests (comprehensive)
2. **comment_service_test.go** - Comment service business logic tests
3. **rating_service_test.go** - Rating service business logic tests
4. **library_service_test.go** - Library service business logic tests

### Repository Tests
To be created based on specific needs - repository tests typically require database integration testing.

## Running Tests

### Run All Tests
```bash
# Run all tests in the entire project
go test ./... -v

# Run all tests with coverage
go test ./... -v -cover

# Run all tests with race detection
go test ./... -race -v
```

### Run Tests by Package

```bash
# Handler tests only
go test ./internal/microservices/http-api/handler -v

# Service tests only
go test ./internal/microservices/http-api/service -v

# Repository tests only (when created)
go test ./internal/microservices/http-api/repository -v
```

### Run Specific Test Files

```bash
# Run auth handler tests
go test ./internal/microservices/http-api/handler -v -run TestAuthHandler

# Run comment service tests
go test ./internal/microservices/http-api/service -v -run TestCommentService

# Run manga handler tests
go test ./internal/microservices/http-api/handler -v -run TestMangaHandler
```

### Run Individual Tests

```bash
# Run specific test function
go test ./internal/microservices/http-api/handler -v -run TestAuthHandler_Login

# Run tests matching pattern
go test ./internal/microservices/http-api/handler -v -run ".*Success"
```

### Coverage Reports

```bash
# Generate coverage report for handlers
go test ./internal/microservices/http-api/handler -coverprofile=coverage_handler.out
go tool cover -html=coverage_handler.out

# Generate coverage report for services
go test ./internal/microservices/http-api/service -coverprofile=coverage_service.out
go tool cover -html=coverage_service.out

# Generate combined coverage
go test ./internal/microservices/http-api/... -coverprofile=coverage_all.out
go tool cover -html=coverage_all.out
```

### Benchmark Tests

```bash
# Run benchmarks if any exist
go test ./internal/microservices/http-api/... -bench=. -benchmem
```

### Continuous Testing

```bash
# Watch for changes and re-run tests (requires gotestsum or similar)
gotestsum --watch ./internal/microservices/http-api/...
```

## Test Structure

All tests follow this structure:
- **Mock Objects**: Created using testify/mock
- **Setup Functions**: Initialize test routers and services
- **Test Cases**: Organized with t.Run() for subtests
- **Assertions**: Use testify/assert for clean assertions
- **Cleanup**: Mock expectations verified with AssertExpectations()

## Mock Middleware

Auth middleware is mocked in tests to simulate authenticated users:

```go
func mockAuthMiddleware(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Set("username", "testuser")
		c.Set("role", role)
		c.Next()
	}
}
```

## Common Test Patterns

### Handler Test
```go
func TestHandler_Action(t *testing.T) {
	mockService := new(MockService)
	r := setupRouter(mockService)

	t.Run("Success", func(t *testing.T) {
		// Setup mock expectations
		mockService.On("Method", args).Return(result, nil)
		
		// Make HTTP request
		req, _ := http.NewRequest("GET", "/endpoint", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
```

### Service Test
```go
func TestService_Method(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	t.Run("Success", func(t *testing.T) {
		// Setup mock expectations
		mockRepo.On("Method", args).Return(result, nil)
		
		// Call service method
		result, err := service.Method(args)
		
		// Assert results
		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockRepo.AssertExpectations(t)
	})
}
```

## Next Steps

1. **Create Repository Tests**: Add integration tests for database operations
2. **Add E2E Tests**: Create end-to-end tests for critical workflows
3. **Increase Coverage**: Aim for 80%+ code coverage
4. **Performance Tests**: Add benchmark tests for critical paths
5. **Edge Cases**: Add more edge case testing

## Dependencies

```bash
# Install testing dependencies if not already installed
go get -u github.com/stretchr/testify/assert
go get -u github.com/stretchr/testify/mock
go get -u github.com/gin-gonic/gin
```

## CI/CD Integration

Add to your CI/CD pipeline:
```yaml
test:
  script:
    - go test ./... -v -cover -race
    - go test ./... -coverprofile=coverage.out
    - go tool cover -func=coverage.out
```
