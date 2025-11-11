# Integration Tests Guide

This document explains how to run integration tests for the MangaHub project.

## Overview

We have created three types of integration tests:

1. **TCP Server Integration Tests** (`tcp_integration_test.go`)
2. **HTTP API Integration Tests** (`http_integration_test.go`)
3. **TCP + HTTP Combined Integration Tests** (`tcp_http_integration_test.go`)

## Prerequisites

### Required Services

Integration tests require the following services to be running:

1. **PostgreSQL Database**
   - Host: `localhost`
   - Port: `5432`
   - Database: `mangahub_test`
   - User: `postgres`
   - Password: `postgres`

2. **Redis Server**
   - Host: `localhost`
   - Port: `6379`
   - No password

### Setup Test Database

```bash
# Create test database
createdb -U postgres mangahub_test

# Or using psql
psql -U postgres
CREATE DATABASE mangahub_test;
\q
```

### Start Redis

```bash
# Using Docker
docker run -d -p 6379:6379 redis:latest

# Or using docker-compose (if you have it)
docker-compose up -d redis
```

## Running Integration Tests

### Run All Integration Tests

```bash
# Run all integration tests
go test -v ./test -run Integration

# Run with timeout
go test -v ./test -run Integration -timeout 5m
```

### Run Specific Integration Tests

#### 1. TCP Server Integration Tests

```bash
# Run all TCP integration tests
go test -v ./test -run TestIntegration_TCPServer

# Run specific TCP test
go test -v ./test -run TestIntegration_TCPServer_ProgressSync
go test -v ./test -run TestIntegration_TCPServer_ErrorRecovery
go test -v ./test -run TestIntegration_TCPServer_HighLoad
go test -v ./test -run TestIntegration_TCPServer_RealWorld
```

#### 2. HTTP API Integration Tests

```bash
# Run all HTTP integration tests
go test -v ./test -run TestIntegration_HTTPServer

# Run specific HTTP test
go test -v ./test -run TestIntegration_HTTPServer_FullAuthFlow
go test -v ./test -run TestIntegration_HTTPServer_MangaCRUD
go test -v ./test -run TestIntegration_HTTPServer_ProgressTracking
```

#### 3. TCP + HTTP Combined Integration Tests

```bash
# Run combined integration tests
go test -v ./test -run TestIntegration_TCPandHTTP

# Run specific combined test
go test -v ./test -run TestIntegration_TCPandHTTP_ProgressSync
go test -v ./test -run TestIntegration_RealWorldScenario
```

### Skip Integration Tests

Integration tests are skipped when using the `-short` flag:

```bash
# Skip integration tests, run only unit tests
go test -v ./test -short
```

## Test Coverage

### TCP Server Integration Tests

| Test | Description | What it validates |
|------|-------------|-------------------|
| **ProgressSync** | Single and multiple client progress updates | - TCP connection<br>- Message broadcast<br>- Redis persistence |
| **ErrorRecovery** | Server resilience to errors | - Malformed message handling<br>- Rapid reconnections<br>- Server stability |
| **HighLoad** | Server under sustained load | - 50 concurrent clients<br>- 10 messages per client<br>- 80% success rate minimum |
| **RealWorld** | Real-world usage patterns | - Reading sessions<br>- Concurrent readers<br>- Progress tracking |

### HTTP API Integration Tests

| Test | Description | What it validates |
|------|-------------|-------------------|
| **FullAuthFlow** | Complete authentication workflow | - User registration<br>- Login<br>- Token refresh<br>- Duplicate prevention |
| **MangaCRUD** | Manga CRUD operations | - Create manga<br>- Read manga<br>- Update manga<br>- Delete manga<br>- List all manga |
| **ProgressTracking** | Progress management | - Update progress<br>- Retrieve progress<br>- List all user progress |

### TCP + HTTP Combined Integration Tests

| Test | Description | What it validates |
|------|-------------|-------------------|
| **ProgressSync** | Cross-service synchronization | - HTTP user registration<br>- HTTP login<br>- HTTP manga creation<br>- TCP progress updates<br>- TCP broadcast<br>- Concurrent TCP+HTTP updates |
| **RealWorldScenario** | Complete user journey | - Sign up (HTTP)<br>- Browse manga (HTTP)<br>- Real-time reading (TCP)<br>- Progress verification (HTTP) |

## Expected Output

### Successful Test Run

```
=== RUN   TestIntegration_TCPServer_ProgressSync
=== RUN   TestIntegration_TCPServer_ProgressSync/SingleClient_ProgressUpdate
    tcp_integration_test.go:61: ✓ Progress update sent and broadcasted successfully
=== RUN   TestIntegration_TCPServer_ProgressSync/MultipleClients_BroadcastSync
    tcp_integration_test.go:114: Client 0 received broadcast
    tcp_integration_test.go:114: Client 1 received broadcast
    tcp_integration_test.go:114: Client 2 received broadcast
    tcp_integration_test.go:118: ✓ Broadcast synchronized across 3 clients
--- PASS: TestIntegration_TCPServer_ProgressSync (1.23s)
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Error

```
Error: failed to connect to test database
```

**Solution:**
- Ensure PostgreSQL is running
- Verify database credentials
- Create `mangahub_test` database
- Check firewall settings

#### 2. Redis Connection Error

```
Error: failed to connect to Redis
```

**Solution:**
- Ensure Redis is running: `redis-cli ping` should return `PONG`
- Check Redis port 6379 is accessible
- Start Redis: `redis-server` or `docker run -p 6379:6379 redis`

#### 3. Port Already in Use

```
Error: bind: address already in use
```

**Solution:**
- Tests use ports 9091-9096 for TCP servers
- Kill processes using these ports
- Or modify test to use different ports

#### 4. Timeout Errors

```
Error: context deadline exceeded
```

**Solution:**
- Increase timeout: `-timeout 10m`
- Check system resources
- Reduce concurrent test load

### Debug Mode

Run tests with verbose logging:

```bash
# Enable debug output
go test -v ./test -run Integration -timeout 5m 2>&1 | tee test-output.log

# Check logs
cat test-output.log
```

## Best Practices

### 1. Clean State Between Tests

Each test should:
- Use unique user IDs/manga IDs
- Clean up database tables
- Close connections properly
- Use `defer` for cleanup

### 2. Isolation

- Tests should be independent
- Don't rely on execution order
- Use separate test databases
- Clear Redis between test suites if needed

### 3. Timeouts

- Set appropriate timeouts for network operations
- Use `context.WithTimeout` for database operations
- Configure test timeout: `-timeout 5m`

### 4. Resource Management

- Close database connections
- Stop TCP servers with `defer`
- Clean up goroutines
- Monitor memory usage

## Performance Benchmarks

### TCP Server

- **Single Client**: < 100ms per message
- **50 Concurrent Clients**: 80%+ success rate
- **Throughput**: > 500 msg/sec under load

### HTTP API

- **Auth Flow**: < 500ms for register+login
- **CRUD Operations**: < 100ms per operation
- **Response Time**: P95 < 200ms

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration-test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: mangahub_test
        ports:
          - 5432:5432
          
      redis:
        image: redis:latest
        ports:
          - 6379:6379
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run Integration Tests
        run: go test -v ./test -run Integration -timeout 5m
```

## Next Steps

1. **Add More Test Cases**
   - Edge cases
   - Error scenarios
   - Load testing variations

2. **Performance Testing**
   - Benchmark critical paths
   - Profile memory usage
   - Test under extreme load

3. **End-to-End Tests**
   - Full system tests
   - Multiple services interaction
   - Failure recovery scenarios

## Summary

✅ **3 Integration Test Suites Created:**
- TCP Server Integration (4 test scenarios)
- HTTP API Integration (3 test scenarios)
- TCP + HTTP Combined Integration (2 test scenarios)

✅ **Total: 9 comprehensive integration tests**

These tests validate:
- Real database interactions
- Redis persistence
- TCP network communication
- HTTP API workflows
- Cross-service synchronization
- Real-world usage patterns

**Run all tests:**
```bash
go test -v ./test -run Integration -timeout 5m
```
