# MangaHub Benchmark Testing Documentation

## Table of Contents

1. [Overview](#overview)
2. [Benchmark Targets](#benchmark-targets)
3. [Test Files](#test-files)
4. [Test Scenarios](#test-scenarios)
5. [How Each Test Works](#how-each-test-works)
6. [Running the Tests](#running-the-tests)
7. [Metrics & Thresholds](#metrics--thresholds)
8. [Architecture Notes](#architecture-notes)
9. [Results Interpretation](#results-interpretation)
10. [Troubleshooting](#troubleshooting)

---

## Overview

The MangaHub benchmark suite provides comprehensive performance testing for the application's microservices architecture. It validates the system against defined performance targets covering:

- **HTTP API** performance under concurrent load
- **TCP Server** connection handling
- **WebSocket** real-time communication
- **Database** query performance
- **Search** latency requirements

### Technology Stack Tested

| Component | Technology | Port |
|-----------|------------|------|
| API Server | Go (Gin) | 8084 |
| TCP Server | Go (net) | 8081 |
| WebSocket | Go (gorilla/websocket) | 8084/ws |
| Database | PostgreSQL + GORM | 5432 |
| Cache | Redis | 6379 |
| Monitoring | Prometheus + Grafana | 9090, 3000 |

---

## Benchmark Targets

### Standard Testing Phase Targets

| Benchmark | Target | Description |
|-----------|--------|-------------|
| Concurrent Users | 50-100 | Simultaneous API requests |
| Database Capacity | 30-40 manga | Records in database |
| Search Latency | <500ms | Query response time |
| TCP Connections | 20-30 | Concurrent TCP clients |
| WebSocket Users | 10-20 | Real-time chat users |
| Uptime | 80-90% | Service availability |

### Stress Testing Targets (Extended)

| Benchmark | Target | Description |
|-----------|--------|-------------|
| HTTP Users | 100-200 | High concurrent load |
| TCP Connections | 100+ | Sustained connections |
| WebSocket Users | 50+ | Real-time messaging |

---

## Test Files

### 1. `benchmark_test.go` (730 lines)

**Purpose:** Go native benchmark tests using the `testing` package.

**Key Features:**
- Type-safe testing with Go's testing framework
- Goroutine-based concurrency for realistic load
- Atomic counters for thread-safe metrics
- WebSocket testing with `gorilla/websocket`

**Test Functions:**
```go
TestConcurrentUsers()      // 100 concurrent API requests
TestSearchLatency()        // Search query performance
TestTCPConnections()       // TCP connection & auth handshake
TestWebSocketConnections() // WebSocket connect & messaging
TestDatabaseCapacity()     // Database record count
TestErrorHandling()        // Error response validation
TestFullBenchmarkSuite()   // Runs all tests together
```

**Configuration:**
```go
type BenchmarkConfig struct {
    APIBaseURL      string        // http://localhost:8084
    TCPAddr         string        // localhost:8081
    WSEndpoint      string        // ws://localhost:8084/ws
    ConcurrentUsers int           // 100
    TCPConnections  int           // 30
    WSConnections   int           // 20
    SearchTimeout   time.Duration // 500ms
}
```

---

### 2. `stress_test.ps1` (464 lines)

**Purpose:** High-load stress testing with real API interactions using PowerShell runspace pools.

**Key Features:**
- **Runspace Pool:** True parallel execution (50 concurrent threads)
- **Real Operations:** Actual API calls (Search, Browse, Detail, Genres)
- **TCP Testing:** Connection establishment and message exchange
- **WebSocket Verification:** Endpoint availability check

**Parameters:**
```powershell
param(
    [int]$HttpUsers = 150,        # Target HTTP concurrent users
    [int]$TcpConnections = 100,   # Target TCP connections
    [int]$WsConnections = 50      # Target WebSocket connections
)
```

**Stress Tests:**

| Test | What It Does | Pass Criteria |
|------|--------------|---------------|
| HTTP Stress | 150-200 concurrent requests with 10 different operations | ≥90% success rate |
| TCP Stress | 100+ connections with message exchange | ≥90% connections established |
| WebSocket | Endpoint availability and upgrade support | Endpoint responds |

**Operations Tested:**
```powershell
$operations = @(
    @{Name="Search"; Endpoint="/api/manga/search?q=naruto"},
    @{Name="Search"; Endpoint="/api/manga/search?q=one+piece"},
    @{Name="Search"; Endpoint="/api/manga/search?q=dragon"},
    @{Name="Browse"; Endpoint="/api/manga/?page=1&limit=20"},
    @{Name="Browse"; Endpoint="/api/manga/?page=2&limit=20"},
    @{Name="Browse"; Endpoint="/api/manga/?page=3&limit=20"},
    @{Name="Detail"; Endpoint="/api/manga/1"},
    @{Name="Detail"; Endpoint="/api/manga/2"},
    @{Name="Detail"; Endpoint="/api/manga/3"},
    @{Name="Genres"; Endpoint="/api/genres/"}
)
```

---

### 3. `quick_validate.ps1` (212 lines)

**Purpose:** Fast one-command validation of all 6 benchmark targets.

**Key Features:**
- Single user authentication
- Sequential test execution
- Clean summary output
- Ideal for quick checks before demos

**Benchmarks Validated:**
1. Search Latency (<500ms) - 10 queries
2. Database Capacity (30-40 manga)
3. Concurrent Users (50 sequential requests)
4. TCP Connections (25 connections)
5. WebSocket Support (endpoint check)
6. Uptime/Reliability (Docker container status)

---

### 4. `demo_benchmarks.ps1` (485 lines)

**Purpose:** Interactive demonstration script with visual progress and explanations.

**Key Features:**
- Color-coded output with progress bars
- Interactive pauses for explanations
- Service health checks before tests
- Suitable for presentations

**Parameters:**
```powershell
param(
    [switch]$Automated,  # Skip pauses for CI/CD
    [switch]$SkipSetup   # Skip health checks
)
```

---

### 5. `run_benchmarks.ps1` (452 lines)

**Purpose:** Comprehensive automated benchmark runner with detailed output.

**Parameters:**
```powershell
param(
    [switch]$Quick,              # Health check only
    [switch]$LoadTest,           # Extended load test
    [switch]$Verbose,            # Detailed output
    [int]$ConcurrentUsers = 100, # Custom user count
    [int]$TCPConnections = 30,   # Custom TCP connections
    [int]$WSConnections = 20     # Custom WS connections
)
```

**Functions:**
```powershell
Test-ServiceHealth()     # Check all services
Test-ConcurrentUsers()   # API concurrent load
Test-SearchLatency()     # Search performance
Test-TCPConnections()    # TCP connection test
Test-DatabaseCapacity()  # Database query
```

---

### 6. `load_test.js` (334 lines)

**Purpose:** k6 load testing script for extended stress scenarios.

**Key Features:**
- Multiple test scenarios running in parallel
- Custom metrics (searchLatency, wsLatency, wsMessages)
- WebSocket stress testing with real messaging
- Configurable thresholds

**Scenarios:**

| Scenario | Executor | VUs | Duration | Description |
|----------|----------|-----|----------|-------------|
| `http_stress` | ramping-vus | 0→200 | 6min | Ramp to 200 concurrent users |
| `search_load` | constant-vus | 30 | 3min | Sustained search queries |
| `websocket_stress` | ramping-vus | 0→75 | 5min | WebSocket connections |
| `spike_test` | ramping-vus | 10→200→10 | 1min | Traffic spike simulation |

**Custom Metrics:**
```javascript
const searchLatency = new Trend('search_latency', true);
const apiLatency = new Trend('api_latency', true);
const wsLatency = new Trend('ws_latency', true);
const errorRate = new Rate('error_rate');
const successfulRequests = new Counter('successful_requests');
const failedRequests = new Counter('failed_requests');
const wsMessages = new Counter('ws_messages');
```

**Thresholds:**
```javascript
thresholds: {
    'http_req_duration': ['p(95)<500'],
    'search_latency': ['p(95)<500', 'avg<300'],
    'ws_latency': ['p(95)<1000'],
    'error_rate': ['rate<0.1'],
    'http_req_failed': ['rate<0.1'],
}
```

---

## How Each Test Works

### HTTP Concurrent Users Test

```
┌─────────────────────────────────────────────────────────────────┐
│                    HTTP CONCURRENT USERS TEST                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. Create test user (or reuse existing)                       │
│     POST /auth/register → POST /auth/login → Get JWT token      │
│                                                                 │
│  2. Launch N concurrent requests (runspace pool)                │
│     ┌──────────┐  ┌──────────┐  ┌──────────┐                   │
│     │ Thread 1 │  │ Thread 2 │  │ Thread N │                   │
│     │ GET /api │  │ GET /api │  │ GET /api │                   │
│     │ /manga/  │  │ /search  │  │ /genres  │                   │
│     └────┬─────┘  └────┬─────┘  └────┬─────┘                   │
│          │             │             │                          │
│          ▼             ▼             ▼                          │
│     ┌─────────────────────────────────────────┐                │
│     │           API Server (8084)             │                │
│     └─────────────────────────────────────────┘                │
│                                                                 │
│  3. Collect results                                            │
│     - Success count, failure count                             │
│     - Latency: avg, p95, max                                   │
│     - Throughput: requests/second                              │
│                                                                 │
│  4. Evaluate                                                   │
│     PASS: ≥90% success rate (stress) or ≥95% (standard)        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### TCP Connection Test

```
┌─────────────────────────────────────────────────────────────────┐
│                    TCP CONNECTION TEST                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  For each connection (1 to N):                                 │
│                                                                 │
│  ┌────────────┐         ┌────────────────────┐                 │
│  │ TCP Client │────────▶│ TCP Server (:8081) │                 │
│  └────────────┘         └────────────────────┘                 │
│        │                         │                              │
│        │  1. TCP Connect         │                              │
│        │─────────────────────────▶                              │
│        │                         │                              │
│        │  2. Send Auth Message   │                              │
│        │  {"type":"auth",        │                              │
│        │   "data":{"token":"x"}} │                              │
│        │─────────────────────────▶                              │
│        │                         │                              │
│        │  3. Receive Response    │                              │
│        │  {"type":"auth_success"}│                              │
│        │◀─────────────────────────                              │
│        │                         │                              │
│        │  4. Send Test Message   │                              │
│        │  {"type":"ping"}        │                              │
│        │─────────────────────────▶                              │
│                                                                 │
│  Metrics Collected:                                            │
│  - Connections established                                     │
│  - Auth success/fail                                           │
│  - Messages sent/received                                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Search Latency Test

```
┌─────────────────────────────────────────────────────────────────┐
│                    SEARCH LATENCY TEST                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Test Queries:                                                  │
│  ["naruto", "one piece", "dragon ball", "attack on titan",     │
│   "my hero academia", "death note", "black clover", ...]       │
│                                                                 │
│  For each query:                                               │
│                                                                 │
│  START TIMER ────────────────────────────────────────┐         │
│       │                                              │         │
│       ▼                                              │         │
│  GET /api/manga/search?q={query}                     │         │
│       │                                              │         │
│       ▼                                              │         │
│  ┌─────────────────────────────────────┐             │         │
│  │ PostgreSQL Full-Text Search         │             │         │
│  │ (GIN index on title, description)   │             │         │
│  └─────────────────────────────────────┘             │         │
│       │                                              │         │
│       ▼                                              │         │
│  Response received                                   │         │
│       │                                              │         │
│  STOP TIMER ─────────────────────────────────────────┘         │
│                                                                 │
│  Metrics:                                                      │
│  - Individual query latency                                    │
│  - Average latency (target: <500ms)                            │
│  - Max latency                                                 │
│  - P95 latency (target: <800ms)                               │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### WebSocket Test

```
┌─────────────────────────────────────────────────────────────────┐
│                    WEBSOCKET TEST                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌────────────┐         ┌────────────────────┐                 │
│  │  WS Client │────────▶│ WS Hub (:8084/ws)  │                 │
│  └────────────┘         └────────────────────┘                 │
│        │                         │                              │
│        │  1. Upgrade Request     │                              │
│        │  Headers:               │                              │
│        │  - Upgrade: websocket   │                              │
│        │  - Connection: Upgrade  │                              │
│        │  - Authorization: Bearer│                              │
│        │─────────────────────────▶                              │
│        │                         │                              │
│        │  2. 101 Switching       │                              │
│        │◀─────────────────────────                              │
│        │                         │                              │
│        │  3. Join Room           │                              │
│        │  {"type":"join",        │                              │
│        │   "room_id": 1}         │                              │
│        │─────────────────────────▶                              │
│        │                         │                              │
│        │  4. Chat Message        │                              │
│        │  {"type":"chat",        │                              │
│        │   "content":"Hello"}    │                              │
│        │─────────────────────────▶                              │
│        │                         │                              │
│        │  5. Broadcast to Room   │                              │
│        │◀─────────────────────────                              │
│                                                                 │
│  Architecture: Hub pattern with goroutine-per-connection       │
│  Capacity: Estimated 1000+ concurrent connections              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Running the Tests

### Quick Validation (Recommended First Step)

```powershell
cd test/benchmark
.\quick_validate.ps1
```

**Output Example:**
```
╔══════════════════════════════════════════════════════════════════════╗
║         MANGAHUB BENCHMARK VALIDATION                                ║
╚══════════════════════════════════════════════════════════════════════╝

[1] SEARCH LATENCY (Target: <500ms)
    'naruto': 45ms ✅ PASS
    'one piece': 52ms ✅ PASS
    📊 Average: 48ms | Pass: 10/10

[2] DATABASE CAPACITY (Target: 30-40 manga)
    Total manga: 140 ✅ EXCEEDS TARGET

...

╔══════════════════════════════════════════════════════════════════════╗
║                       BENCHMARK SUMMARY                              ║
║  Search (<500ms)         │ 48 ms avg        │ ✅ PASS               ║
║  Database (30-40 manga)  │ 140 manga        │ ✅ PASS               ║
║  Concurrent Users (50+)  │ 50/50 @ 45req/s  │ ✅ PASS               ║
║  TCP Connections (20-30) │ 25/25 connections│ ✅ PASS               ║
║  WebSocket (10-20 users) │ Hub pattern      │ ✅ PASS               ║
║  Uptime (80-90%)         │ 14 containers    │ ✅ PASS               ║
╚══════════════════════════════════════════════════════════════════════╝

  ✅ ALL BENCHMARKS PASSED
```

### Stress Testing (High Load)

```powershell
# Default: 150 HTTP, 100 TCP, 50 WebSocket
.\stress_test.ps1

# Custom targets
.\stress_test.ps1 -HttpUsers 200 -TcpConnections 120 -WsConnections 75
```

**Output Example:**
```
╔══════════════════════════════════════════════════════════════════════╗
║         MANGAHUB STRESS TEST - REAL INTERACTIONS                     ║
║  HTTP Users:    200   │  TCP Connections: 120                        ║
╚══════════════════════════════════════════════════════════════════════╝

  ┌────────────────────────────────────────────────────────────┐
  │  HTTP STRESS TEST RESULTS                                  │
  │  Total Requests:    200                                    │
  │  Successful:        180        (90%)                       │
  │  Throughput:        172.4      req/sec                     │
  │  Avg Latency:       3.49       ms                          │
  │  P95 Latency:       9          ms                          │
  └────────────────────────────────────────────────────────────┘

  Result: ✅ PASS - 90% success rate at 172.4 req/sec
```

### Go Native Tests

```powershell
# All benchmarks
go test -v -timeout 10m ./test/benchmark/...

# Specific test
go test -v -run TestConcurrentUsers ./test/benchmark/...
go test -v -run TestSearchLatency ./test/benchmark/...
go test -v -run TestTCPConnections ./test/benchmark/...
go test -v -run TestWebSocketConnections ./test/benchmark/...

# Full suite
go test -v -run TestFullBenchmarkSuite ./test/benchmark/...
```

### k6 Load Testing

```powershell
# Install k6: https://k6.io/docs/get-started/installation/
# Or: choco install k6

# Run full load test (all scenarios)
k6 run test/benchmark/load_test.js

# Custom VUs and duration
k6 run --vus 100 --duration 5m test/benchmark/load_test.js

# Output to JSON for analysis
k6 run --out json=results.json test/benchmark/load_test.js

# Run specific scenario
k6 run --env SCENARIO=http_stress test/benchmark/load_test.js
```

### Interactive Demo (Presentations)

```powershell
.\demo_benchmarks.ps1           # Interactive with pauses
.\demo_benchmarks.ps1 -Automated # No pauses for CI/CD
```

---

## Metrics & Thresholds

### HTTP Metrics

| Metric | Description | Target |
|--------|-------------|--------|
| Success Rate | % of 2xx responses | ≥90% (stress), ≥95% (standard) |
| Avg Latency | Mean response time | <100ms |
| P95 Latency | 95th percentile | <500ms |
| Max Latency | Worst case | <2000ms |
| Throughput | Requests per second | >100 rps |

### Search Metrics

| Metric | Description | Target |
|--------|-------------|--------|
| Avg Latency | Mean search time | <300ms |
| P95 Latency | 95th percentile | <500ms |
| Max Latency | Worst case | <800ms |

### TCP Metrics

| Metric | Description | Target |
|--------|-------------|--------|
| Connection Rate | % connections established | ≥80% |
| Auth Success | % authenticated successfully | ≥80% |
| Stability | Connections alive after 2s | ≥90% |

### WebSocket Metrics

| Metric | Description | Target |
|--------|-------------|--------|
| Connection Rate | % WebSocket upgrades | ≥80% |
| Message Latency | Time to receive broadcast | <1000ms |
| Messages/sec | Throughput | >10 msg/s/client |

---

## Architecture Notes

### Concurrency Model

```
┌─────────────────────────────────────────────────────────────────┐
│                    MangaHub Concurrency Model                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  HTTP (Gin Framework)                                          │
│  ├── Connection pool to PostgreSQL                             │
│  ├── Redis caching layer                                       │
│  └── Goroutine per request (managed by Gin)                    │
│                                                                 │
│  TCP Server                                                    │
│  ├── Accept loop in main goroutine                             │
│  ├── Goroutine per connection                                  │
│  ├── JSON message protocol                                     │
│  └── Authentication via JWT                                    │
│                                                                 │
│  WebSocket (Hub Pattern)                                       │
│  ├── Central Hub manages all connections                       │
│  ├── Goroutine per connection (read/write pumps)               │
│  ├── Channel-based message broadcasting                        │
│  └── Room-based subscription                                   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Connection Flow

```
Client Request → Load Balancer → API Server → Database/Cache → Response
                                     │
                                     ├── PostgreSQL (persistent data)
                                     └── Redis (session, cache)
```

---

## Results Interpretation

### Success Indicators

✅ **All Benchmarks Passed** when:
- HTTP success rate ≥90% under load
- Search avg latency <500ms
- TCP connections ≥80% established
- WebSocket endpoint responds
- Database has ≥30 records

### Warning Signs

⚠️ **Investigation Needed** when:
- Success rate 80-90%
- Latency approaching thresholds
- Connection timeouts occurring
- High P95/max latency variance

### Failure Indicators

❌ **Action Required** when:
- Success rate <80%
- Average latency >500ms for search
- Connections failing >20%
- Server errors (5xx) appearing

---

## Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Auth failures | Token expired | Re-login, check JWT expiry |
| 301 Redirects | Missing trailing slash | Add `/` to API URLs |
| Connection refused | Service not running | `docker-compose up -d` |
| Timeout errors | Server overloaded | Reduce concurrent load |
| Low success rate | Resource exhaustion | Check container resources |

### Performance Tuning

```yaml
# docker-compose.yml adjustments
services:
  api:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
```

```go
// database/db.go - Connection pool tuning
sqlDB.SetMaxOpenConns(100)
sqlDB.SetMaxIdleConns(25)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

### Debug Commands

```powershell
# Check container status
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# View API logs
docker logs mangahub-api -f --tail 100

# Check database connections
docker exec mangahub-postgres psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

# Monitor Redis
docker exec mangahub-redis redis-cli INFO clients
```

---

## File Summary

| File | Lines | Purpose | Usage |
|------|-------|---------|-------|
| `benchmark_test.go` | 730 | Go native benchmarks | `go test -v ./test/benchmark/...` |
| `stress_test.ps1` | 464 | High-load stress test | `.\stress_test.ps1 -HttpUsers 200` |
| `quick_validate.ps1` | 212 | Fast validation | `.\quick_validate.ps1` |
| `demo_benchmarks.ps1` | 485 | Interactive demo | `.\demo_benchmarks.ps1` |
| `run_benchmarks.ps1` | 452 | Automated runner | `.\run_benchmarks.ps1` |
| `load_test.js` | 334 | k6 load testing | `k6 run load_test.js` |
| `README.md` | 336 | Quick reference | - |

---

## Quick Reference Commands

```powershell
# Prerequisites
docker-compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d

# Quick check (1 minute)
.\quick_validate.ps1

# Standard benchmarks (5 minutes)
.\run_benchmarks.ps1

# Stress test (2 minutes)
.\stress_test.ps1 -HttpUsers 200 -TcpConnections 120

# Full Go test suite (10 minutes)
go test -v -timeout 10m ./test/benchmark/...

# Extended k6 load test (8 minutes)
k6 run test/benchmark/load_test.js

# Demo for presentations
.\demo_benchmarks.ps1
```

---

*Documentation generated: December 16, 2025*
*MangaHub Benchmark Suite v1.0*
