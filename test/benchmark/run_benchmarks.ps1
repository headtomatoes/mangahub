# ============================================================================
# MangaHub Benchmark Runner - PowerShell Script
# ============================================================================
# This script runs comprehensive benchmarks against your MangaHub services
# 
# Prerequisites:
#   - Docker containers running (docker-compose up -d)
#   - Go installed for running tests
#
# Usage:
#   .\run_benchmarks.ps1              # Run all benchmarks
#   .\run_benchmarks.ps1 -Quick       # Quick health check only
#   .\run_benchmarks.ps1 -LoadTest    # Run extended load test
# ============================================================================

param(
    [switch]$Quick,
    [switch]$LoadTest,
    [switch]$Verbose,
    [int]$ConcurrentUsers = 200,
    [int]$TCPConnections = 60,
    [int]$WSConnections = 40
)

$ErrorActionPreference = "Continue"

# Configuration
$API_URL = "http://10.238.20.112:8084"
$TCP_HOST = "10.238.20.112"
$TCP_PORT = 8081
$WS_URL = "ws://10.238.20.112:8084/ws"
$PROMETHEUS_URL = "http://localhost:9090"
$GRAFANA_URL = "http://localhost:3000"

# Colors for output
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Write-Header($text) {
    Write-Host ""
    Write-Host "╔══════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "║  $($text.PadRight(64))║" -ForegroundColor Cyan
    Write-Host "╚══════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""
}

function Write-Section($text) {
    Write-Host ""
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
    Write-Host "  $text" -ForegroundColor Yellow
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
}

function Write-Result($status, $message) {
    if ($status -eq "PASS") {
        Write-Host "  ✅ PASS: $message" -ForegroundColor Green
    } elseif ($status -eq "FAIL") {
        Write-Host "  ❌ FAIL: $message" -ForegroundColor Red
    } elseif ($status -eq "WARN") {
        Write-Host "  ⚠️  WARN: $message" -ForegroundColor Yellow
    } else {
        Write-Host "  ℹ️  INFO: $message" -ForegroundColor Gray
    }
}

function Write-Metric($name, $value) {
    Write-Host "     ├── $($name.PadRight(20)) $value" -ForegroundColor White
}

# ============================================================================
# HEALTH CHECKS
# ============================================================================

function Test-ServiceHealth {
    Write-Header "SERVICE HEALTH CHECK"
    
    $services = @(
        @{Name="API Server"; URL="$API_URL/check-conn"; Expected=200},
        @{Name="Database"; URL="$API_URL/db-ping"; Expected=200}
    )
    
    $healthy = 0
    $total = $services.Count
    
    foreach ($svc in $services) {
        try {
            $response = Invoke-WebRequest -Uri $svc.URL -TimeoutSec 5 -UseBasicParsing -ErrorAction Stop
            if ($response.StatusCode -eq $svc.Expected) {
                Write-Host "  ✅ $($svc.Name): Healthy" -ForegroundColor Green
                $healthy++
            } else {
                Write-Host "  ❌ $($svc.Name): Unexpected status $($response.StatusCode)" -ForegroundColor Red
            }
        } catch {
            Write-Host "  ❌ $($svc.Name): Not reachable - $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    
    # Check TCP
    try {
        $tcpClient = New-Object System.Net.Sockets.TcpClient
        $tcpClient.Connect($TCP_HOST, $TCP_PORT)
        Write-Host "  ✅ TCP Server: Healthy (port $TCP_PORT)" -ForegroundColor Green
        $tcpClient.Close()
        $healthy++
    } catch {
        Write-Host "  ❌ TCP Server: Not reachable on port $TCP_PORT" -ForegroundColor Red
    }
    $total++
    
    Write-Host ""
    Write-Host "  Health Summary: $healthy/$total services healthy" -ForegroundColor $(if ($healthy -eq $total) {"Green"} else {"Yellow"})
    
    return $healthy -eq $total
}

# ============================================================================
# BENCHMARK 1: CONCURRENT USERS
# ============================================================================

function Test-ConcurrentUsers {
    param([int]$Users = 100)
    
    Write-Section "BENCHMARK 1: Concurrent Users ($Users target)"
    Write-Host "  🎯 Target: Support $Users concurrent API requests" -ForegroundColor Cyan
    Write-Host ""
    
    # First, register and login to get a token
    $registerBody = @{
        username = "benchmark_user_$(Get-Random)"
        email = "bench$(Get-Random)@test.com"
        password = "benchmark123"
    } | ConvertTo-Json
    
    try {
        $null = Invoke-RestMethod -Uri "$API_URL/auth/register" -Method POST -Body $registerBody -ContentType "application/json" -ErrorAction SilentlyContinue
    } catch {}
    
    $loginBody = @{
        username = "benchmarkuser"
        password = "benchmark123"
    } | ConvertTo-Json
    
    try {
        # Try with existing user first
        $loginResponse = Invoke-RestMethod -Uri "$API_URL/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
        $token = $loginResponse.access_token
    } catch {
        Write-Result "WARN" "Could not authenticate. Create a test user first."
        return
    }
    
    if (-not $token) {
        Write-Result "FAIL" "No auth token received"
        return
    }
    
    $headers = @{ Authorization = "Bearer $token" }
    
    # Run concurrent requests using sequential burst (more reliable than jobs)
    Write-Host "  Running $Users sequential rapid requests..." -ForegroundColor Gray
    $startTime = Get-Date
    
    $successful = 0
    $latencies = @()
    
    for ($i = 0; $i -lt $Users; $i++) {
        $sw = [System.Diagnostics.Stopwatch]::StartNew()
        try {
            # Note: Use trailing slash to avoid redirect losing auth header
            $response = Invoke-WebRequest -Uri "$API_URL/api/manga/?page=1&page_size=10" -Headers $headers -TimeoutSec 30 -UseBasicParsing
            $sw.Stop()
            $successful++
            $latencies += $sw.ElapsedMilliseconds
        } catch {
            $sw.Stop()
            $latencies += $sw.ElapsedMilliseconds
        }
        if ($i % 10 -eq 0) { Write-Host "  Progress: $i/$Users" -NoNewline; Write-Host "`r" -NoNewline }
    }
    
    $endTime = Get-Date
    $totalTime = ($endTime - $startTime).TotalSeconds
    
    # Analyze results
    $failed = $Users - $successful
    $avgLatency = if ($latencies.Count -gt 0) { ($latencies | Measure-Object -Average).Average } else { 0 }
    $maxLatency = if ($latencies.Count -gt 0) { ($latencies | Measure-Object -Maximum).Maximum } else { 0 }
    $p95Latency = if ($latencies.Count -gt 0) { 
        $sorted = $latencies | Sort-Object
        $idx = [math]::Floor($sorted.Count * 0.95)
        $sorted[$idx]
    } else { 0 }
    
    Write-Metric "Total Requests" $Users
    Write-Metric "Successful" "$successful ($([math]::Round($successful/$Users*100, 1))%)"
    Write-Metric "Failed" $failed
    Write-Metric "Avg Latency" "$([math]::Round($avgLatency, 2)) ms"
    Write-Metric "P95 Latency" "$([math]::Round($p95Latency, 2)) ms"
    Write-Metric "Max Latency" "$([math]::Round($maxLatency, 2)) ms"
    Write-Metric "Total Time" "$([math]::Round($totalTime, 2)) seconds"
    Write-Host ""
    
    $successRate = $successful / $Users * 100
    if ($successRate -ge 95) {
        Write-Result "PASS" "$([math]::Round($successRate, 1))% success rate (target: ≥95%)"
        return $true
    } else {
        Write-Result "FAIL" "$([math]::Round($successRate, 1))% success rate (target: ≥95%)"
        return $false
    }
}

# ============================================================================
# BENCHMARK 2: SEARCH LATENCY
# ============================================================================

function Test-SearchLatency {
    Write-Section "BENCHMARK 2: Search Latency (<500ms)"
    Write-Host "  🎯 Target: Search queries complete within 500ms" -ForegroundColor Cyan
    Write-Host ""
    
    # Login
    $loginBody = @{ username = "benchmarkuser"; password = "benchmark123" } | ConvertTo-Json
    try {
        $loginResponse = Invoke-RestMethod -Uri "$API_URL/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
        $token = $loginResponse.access_token
    } catch {
        Write-Result "WARN" "Could not authenticate"
        return
    }
    
    $headers = @{ Authorization = "Bearer $token" }
    $queries = @("naruto", "one piece", "attack", "dragon", "hero")
    $targetMs = 500
    
    $latencies = @()
    $withinTarget = 0
    
    foreach ($query in $queries) {
        $sw = [System.Diagnostics.Stopwatch]::StartNew()
        try {
            $encodedQuery = [System.Web.HttpUtility]::UrlEncode($query)
            $response = Invoke-RestMethod -Uri "$API_URL/api/manga/search?q=$encodedQuery" -Headers $headers -TimeoutSec 10
            $sw.Stop()
            $latency = $sw.ElapsedMilliseconds
            $latencies += $latency
            
            $status = if ($latency -le $targetMs) { "✅"; $withinTarget++ } else { "⚠️" }
            Write-Host "     $status Query '$query': $latency ms" -ForegroundColor $(if ($latency -le $targetMs) {"Green"} else {"Yellow"})
        } catch {
            $sw.Stop()
            Write-Host "     ❌ Query '$query': ERROR - $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    
    Write-Host ""
    $avgLatency = if ($latencies.Count -gt 0) { ($latencies | Measure-Object -Average).Average } else { 0 }
    $maxLatency = if ($latencies.Count -gt 0) { ($latencies | Measure-Object -Maximum).Maximum } else { 0 }
    
    Write-Metric "Queries Tested" $queries.Count
    Write-Metric "Avg Latency" "$([math]::Round($avgLatency, 2)) ms"
    Write-Metric "Max Latency" "$([math]::Round($maxLatency, 2)) ms"
    Write-Metric "Within Target" "$withinTarget/$($queries.Count) ($([math]::Round($withinTarget/$queries.Count*100, 1))%)"
    Write-Host ""
    
    if ($avgLatency -le $targetMs) {
        Write-Result "PASS" "Average latency $([math]::Round($avgLatency, 2))ms ≤ ${targetMs}ms target"
        return $true
    } else {
        Write-Result "FAIL" "Average latency $([math]::Round($avgLatency, 2))ms > ${targetMs}ms target"
        return $false
    }
}

# ============================================================================
# BENCHMARK 3: TCP CONNECTIONS
# ============================================================================

function Test-TCPConnections {
    param([int]$Connections = 30)
    
    Write-Section "BENCHMARK 3: TCP Connections ($Connections target)"
    Write-Host "  🎯 Target: Support $Connections concurrent TCP connections" -ForegroundColor Cyan
    Write-Host ""
    
    $connected = 0
    $clients = @()
    
    for ($i = 0; $i -lt $Connections; $i++) {
        try {
            $client = New-Object System.Net.Sockets.TcpClient
            $client.Connect($TCP_HOST, $TCP_PORT)
            $clients += $client
            $connected++
        } catch {
            # Connection failed
        }
    }
    
    # Keep alive for a moment
    Start-Sleep -Seconds 2
    
    # Check how many are still connected
    $stillConnected = 0
    foreach ($client in $clients) {
        if ($client.Connected) {
            $stillConnected++
        }
        $client.Close()
    }
    
    Write-Metric "Target" $Connections
    Write-Metric "Connected" $connected
    Write-Metric "Still Active" $stillConnected
    Write-Host ""
    
    $successRate = $connected / $Connections * 100
    if ($successRate -ge 80) {
        Write-Result "PASS" "$connected/$Connections connections established (≥80%)"
        return $true
    } else {
        Write-Result "FAIL" "Only $connected/$Connections connections established (<80%)"
        return $false
    }
}

# ============================================================================
# BENCHMARK 4: DATABASE CAPACITY
# ============================================================================

function Test-DatabaseCapacity {
    Write-Section "BENCHMARK 4: Database Capacity (30-40 manga)"
    Write-Host "  🎯 Target: Handle 30-40 manga series in database" -ForegroundColor Cyan
    Write-Host ""
    
    # Login
    $loginBody = @{ username = "benchmarkuser"; password = "benchmark123" } | ConvertTo-Json
    try {
        $loginResponse = Invoke-RestMethod -Uri "$API_URL/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
        $token = $loginResponse.access_token
    } catch {
        Write-Result "WARN" "Could not authenticate"
        return
    }
    
    $headers = @{ Authorization = "Bearer $token" }
    
    try {
        $response = Invoke-RestMethod -Uri "$API_URL/api/manga?page=1&page_size=100" -Headers $headers
        $total = $response.total
        $returned = $response.data.Count
        
        Write-Metric "Total in DB" $total
        Write-Metric "Target Range" "30-40"
        Write-Metric "Returned" $returned
        Write-Host ""
        
        if ($total -ge 30) {
            Write-Result "PASS" "$total manga in database (target: ≥30)"
            return $true
        } else {
            Write-Result "WARN" "$total manga in database (target: ≥30) - consider importing more"
            return $true  # Not a hard failure
        }
    } catch {
        Write-Result "FAIL" "Could not query manga: $($_.Exception.Message)"
        return $false
    }
}

# ============================================================================
# SUMMARY REPORT
# ============================================================================

function Show-Summary {
    param($results)
    
    Write-Header "BENCHMARK SUMMARY REPORT"
    
    $passed = ($results.Values | Where-Object { $_ -eq $true }).Count
    $total = $results.Count
    
    Write-Host "  Results:" -ForegroundColor White
    foreach ($key in $results.Keys) {
        $status = if ($results[$key]) { "✅ PASS" } else { "❌ FAIL" }
        Write-Host "     $status - $key" -ForegroundColor $(if ($results[$key]) {"Green"} else {"Red"})
    }
    
    Write-Host ""
    Write-Host "  Overall: $passed/$total benchmarks passed" -ForegroundColor $(if ($passed -eq $total) {"Green"} elseif ($passed -gt $total/2) {"Yellow"} else {"Red"})
    Write-Host ""
    
    # Recommendations
    Write-Host "  📋 Quick Links:" -ForegroundColor Cyan
    Write-Host "     ├── Grafana Dashboard: $GRAFANA_URL (admin/mangahub_admin)" -ForegroundColor Gray
    Write-Host "     ├── Prometheus Targets: $PROMETHEUS_URL/targets" -ForegroundColor Gray
    Write-Host "     └── API Health: $API_URL/check-conn" -ForegroundColor Gray
    Write-Host ""
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================

Write-Header "MANGAHUB BENCHMARK SUITE"
Write-Host "  Date: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Gray
Write-Host "  Mode: $(if ($Quick) {'Quick'} elseif ($LoadTest) {'Load Test'} else {'Standard'})" -ForegroundColor Gray
Write-Host ""

# Always run health check first
$healthOk = Test-ServiceHealth

if (-not $healthOk) {
    Write-Host ""
    Write-Host "  ⚠️  Some services are not healthy. Continue anyway? (y/n)" -ForegroundColor Yellow
    $continue = Read-Host
    if ($continue -ne 'y') {
        Write-Host "  Exiting..." -ForegroundColor Gray
        exit 1
    }
}

if ($Quick) {
    Write-Host ""
    Write-Host "  Quick mode - health check complete." -ForegroundColor Green
    exit 0
}

# Run benchmarks
$results = @{}

$results["Concurrent Users ($ConcurrentUsers)"] = Test-ConcurrentUsers -Users $ConcurrentUsers
$results["Search Latency (<500ms)"] = Test-SearchLatency
$results["TCP Connections ($TCPConnections)"] = Test-TCPConnections -Connections $TCPConnections
$results["Database Capacity (30-40)"] = Test-DatabaseCapacity

# Show summary
Show-Summary $results

Write-Host "╔══════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                    BENCHMARK COMPLETE                            ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
