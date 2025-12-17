# MangaHub Benchmark Testing Guide

## 📋 Overview

This directory contains comprehensive benchmark tools to validate MangaHub's performance against the testing phase targets:

| Benchmark | Target | Tool |
|-----------|--------|------|
| Concurrent Users | 50-100 | Go Tests, PowerShell, k6 |
| Database Capacity | 30-40 manga | Go Tests, PowerShell |
| Search Latency | <500ms | Go Tests, k6 |
| TCP Connections | 20-30 | Go Tests, PowerShell |
| WebSocket Connections | 10-20 | Go Tests |
| Uptime/Reliability | 80-90% | Prometheus + Grafana |

---

## 🚀 Quick Start

### Prerequisites

1. **Services Running**
   ```powershell
   # Start all services including monitoring
   docker-compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d
   ```

2. **Create Test User** (optional - scripts will create one)
   ```powershell
   # Via curl or Postman
   curl -X POST http://localhost:8084/auth/register `
     -H "Content-Type: application/json" `
     -d '{"username":"benchmarkuser","email":"bench@test.com","password":"benchmark123"}'
   ```

---

## 🧪 Testing Options

### Option 1: Interactive Demo (Recommended for Presentations)

```powershell
cd test/benchmark
.\demo_benchmarks.ps1
```

This provides a visual, step-by-step demonstration of each benchmark with:
- Live progress indicators
- Color-coded results
- Interactive pauses for explanation

### Option 2: Automated Benchmark Suite

```powershell
cd test/benchmark
.\run_benchmarks.ps1
```

Options:
```powershell
.\run_benchmarks.ps1 -Quick              # Health check only
.\run_benchmarks.ps1 -ConcurrentUsers 50 # Custom user count
.\run_benchmarks.ps1 -TCPConnections 20  # Custom TCP connections
```

### Option 3: Go Test Suite

```powershell
# Run all benchmarks
go test -v -timeout 10m ./test/benchmark/...

# Run specific benchmark
go test -v -run TestConcurrentUsers ./test/benchmark/...
go test -v -run TestSearchLatency ./test/benchmark/...
go test -v -run TestTCPConnections ./test/benchmark/...
go test -v -run TestWebSocketConnections ./test/benchmark/...

# Run full suite
go test -v -run TestFullBenchmarkSuite ./test/benchmark/...
```

### Option 4: k6 Load Testing (Advanced)

```powershell
# Install k6 first: https://k6.io/docs/get-started/installation/
# Or via Chocolatey: choco install k6

# Run load test
k6 run test/benchmark/load_test.js

# With custom settings
k6 run --vus 50 --duration 2m test/benchmark/load_test.js

# Output results to JSON
k6 run --out json=results.json test/benchmark/load_test.js
```

---

## 📊 Benchmark Details

### 1. Concurrent Users (50-100)

**What it tests:**
- API server's ability to handle simultaneous requests
- Connection pooling effectiveness
- Response time under load

**Pass criteria:**
- ≥95% success rate
- No connection refused errors
- P95 latency <2s

**Files:**
- `benchmark_test.go` → `TestConcurrentUsers`
- `run_benchmarks.ps1` → `Test-ConcurrentUsers`
- `load_test.js` → `concurrent_users` scenario

---

### 2. Search Latency (<500ms)

**What it tests:**
- Database query performance
- Index effectiveness
- Full-text search speed

**Pass criteria:**
- Average latency <500ms
- P95 latency <800ms
- No timeouts

**Files:**
- `benchmark_test.go` → `TestSearchLatency`
- `run_benchmarks.ps1` → `Test-SearchLatency`
- `load_test.js` → `searchScenario`

---

### 3. TCP Connections (20-30)

**What it tests:**
- TCP server connection handling
- Goroutine-per-connection pattern
- Connection stability

**Pass criteria:**
- ≥80% connections established
- Connections remain stable for 2+ seconds
- Proper authentication handshake

**Files:**
- `benchmark_test.go` → `TestTCPConnections`
- `run_benchmarks.ps1` → `Test-TCPConnections`

---

### 4. WebSocket Connections (10-20)

**What it tests:**
- WebSocket upgrade handling
- Hub message broadcasting
- Room join/leave functionality

**Pass criteria:**
- ≥80% connections established
- Messages can be sent/received
- Proper room management

**Files:**
- `benchmark_test.go` → `TestWebSocketConnections`

---

### 5. Database Capacity (30-40 manga)

**What it tests:**
- Current data in database
- Pagination performance
- Data retrieval speed

**Pass criteria:**
- ≥30 manga records
- Pagination works correctly
- Response includes total count

**Files:**
- `benchmark_test.go` → `TestDatabaseCapacity`
- `run_benchmarks.ps1` → `Test-DatabaseCapacity`

---

## 📈 Monitoring Integration

### Grafana Dashboard

Access Grafana at `http://localhost:3000` (admin/mangahub_admin)

Pre-configured dashboards show:
- Request rates
- Latency percentiles
- Error rates
- Connection counts

### Prometheus Metrics

Access Prometheus at `http://localhost:9090`

Key metrics:
- `http_request_duration_seconds` - API latency
- `tcp_connections_active` - Active TCP connections
- `ws_connections_active` - WebSocket connections
- `go_goroutines` - Goroutine count

### Alertmanager

Access at `http://localhost:9093`

Configured alerts:
- High error rate (>5%)
- High latency (>500ms p95)
- Service down

---

## 🔧 Troubleshooting

### "Services not healthy"

```powershell
# Check container status
docker-compose ps

# View logs
docker-compose logs api-server
docker-compose logs tcp-server

# Restart services
docker-compose restart
```

### "Authentication failed"

```powershell
# Manually create test user
$body = @{
    username = "benchmarkuser"
    email = "bench@test.com"
    password = "benchmark123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8084/auth/register" `
    -Method POST -Body $body -ContentType "application/json"
```

### "TCP connections timing out"

```powershell
# Check TCP server is running
Test-NetConnection -ComputerName localhost -Port 8081

# Check firewall
Get-NetFirewallRule | Where-Object { $_.LocalPort -eq 8081 }
```

### "k6 not found"

```powershell
# Install via Chocolatey
choco install k6

# Or download from https://k6.io/docs/get-started/installation/
```

---

## 📁 File Structure

```
test/benchmark/
├── README.md                 # This file
├── benchmark_test.go         # Go test suite
├── run_benchmarks.ps1        # PowerShell automated runner
├── demo_benchmarks.ps1       # Interactive demo script
└── load_test.js              # k6 load testing script
```

---

## 🎯 Expected Results (Testing Phase)

For a healthy system in testing phase, you should see:

```
╔══════════════════════════════════════════════════════════════════╗
║                    BENCHMARK SUMMARY                             ║
╚══════════════════════════════════════════════════════════════════╝

  ✅ PASS - Concurrent Users (100): 97.5% success rate
  ✅ PASS - Search Latency: 245ms avg (target: <500ms)
  ✅ PASS - TCP Connections (30): 100% established
  ✅ PASS - WebSocket Connections (20): 95% established
  ✅ PASS - Database Capacity: 42 manga (target: ≥30)
  ✅ PASS - Error Handling: All tests passed

  Overall: 6/6 benchmarks passed
```

---

## 📝 Adding Custom Benchmarks

To add a new benchmark:

1. **Go Test:** Add a new `TestXxx` function to `benchmark_test.go`
2. **PowerShell:** Add a new `Test-Xxx` function to `run_benchmarks.ps1`
3. **k6:** Add a new scenario to `load_test.js`

Example template:
```go
func TestNewBenchmark(t *testing.T) {
    cfg := DefaultConfig()
    
    t.Log("🎯 Target: Description")
    t.Log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    
    // Your test logic here
    
    if passed {
        t.Logf("✅ PASS: Message")
    } else {
        t.Errorf("❌ FAIL: Message")
    }
}
```
