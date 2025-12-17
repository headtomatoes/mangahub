# ============================================================================
# MangaHub Benchmark Demo Script
# ============================================================================
# This script provides an interactive demonstration of all benchmark targets
# with visual output and real-time metrics
#
# Usage:
#   .\demo_benchmarks.ps1                 # Full interactive demo
#   .\demo_benchmarks.ps1 -Automated      # Non-interactive for CI/CD
# ============================================================================

param(
    [switch]$Automated,
    [switch]$SkipSetup
)

$ErrorActionPreference = "Continue"

# Configuration
$API_URL = "http://localhost:8084"
$TCP_HOST = "localhost"
$TCP_PORT = 8081
$PROMETHEUS_URL = "http://localhost:9090"
$GRAFANA_URL = "http://localhost:3000"

# Test user
$TEST_USER = @{
    username = "demouser"
    password = "demo123456"
    email = "demo@mangahub.test"
}

$Global:AuthToken = $null

# ============================================================================
# UTILITY FUNCTIONS
# ============================================================================

function Write-Banner {
    param([string]$text)
    $width = 70
    $padding = [math]::Max(0, ($width - $text.Length - 4) / 2)
    $paddingStr = " " * [math]::Floor($padding)
    
    Write-Host ""
    Write-Host ("╔" + "═" * ($width - 2) + "╗") -ForegroundColor Cyan
    Write-Host ("║" + $paddingStr + $text + $paddingStr + (" " * (($width - 2) - $paddingStr.Length * 2 - $text.Length)) + "║") -ForegroundColor Cyan
    Write-Host ("╚" + "═" * ($width - 2) + "╝") -ForegroundColor Cyan
    Write-Host ""
}

function Write-Section {
    param([string]$title, [string]$target)
    Write-Host ""
    Write-Host ("━" * 60) -ForegroundColor DarkGray
    Write-Host "  📊 $title" -ForegroundColor Yellow
    Write-Host "  🎯 Target: $target" -ForegroundColor Cyan
    Write-Host ("━" * 60) -ForegroundColor DarkGray
    Write-Host ""
}

function Write-Progress-Bar {
    param([int]$current, [int]$total, [string]$label)
    $percent = [math]::Round(($current / $total) * 100)
    $filled = [math]::Round(($current / $total) * 40)
    $empty = 40 - $filled
    $bar = ("█" * $filled) + ("░" * $empty)
    Write-Host "`r  [$bar] $percent% - $label" -NoNewline
}

function Write-Result {
    param([bool]$passed, [string]$metric, [string]$value, [string]$target)
    $icon = if ($passed) { "✅" } else { "❌" }
    $color = if ($passed) { "Green" } else { "Red" }
    Write-Host "  $icon $metric`: " -NoNewline
    Write-Host "$value" -ForegroundColor $color -NoNewline
    Write-Host " (target: $target)"
}

function Pause-Demo {
    if (-not $Automated) {
        Write-Host ""
        Write-Host "  Press Enter to continue..." -ForegroundColor DarkGray
        $null = Read-Host
    } else {
        Start-Sleep -Seconds 2
    }
}

# ============================================================================
# SETUP
# ============================================================================

function Initialize-Demo {
    Write-Banner "MANGAHUB BENCHMARK DEMONSTRATION"
    
    Write-Host "  📋 Benchmark Targets:" -ForegroundColor White
    Write-Host "     ├── 50-100 concurrent users" -ForegroundColor Gray
    Write-Host "     ├── 30-40 manga series in database" -ForegroundColor Gray
    Write-Host "     ├── Search queries < 500ms" -ForegroundColor Gray
    Write-Host "     ├── 20-30 concurrent TCP connections" -ForegroundColor Gray
    Write-Host "     └── 10-20 WebSocket chat users" -ForegroundColor Gray
    Write-Host ""
    
    # Check services
    Write-Host "  🔍 Checking services..." -ForegroundColor Yellow
    
    $services = @(
        @{Name="API Server"; URL="$API_URL/check-conn"},
        @{Name="Database"; URL="$API_URL/db-ping"},
        @{Name="Prometheus"; URL="$PROMETHEUS_URL/-/healthy"},
        @{Name="Grafana"; URL="$GRAFANA_URL/api/health"}
    )
    
    $allHealthy = $true
    foreach ($svc in $services) {
        try {
            $response = Invoke-WebRequest -Uri $svc.URL -TimeoutSec 3 -UseBasicParsing -ErrorAction Stop
            Write-Host "     ✅ $($svc.Name)" -ForegroundColor Green
        } catch {
            Write-Host "     ❌ $($svc.Name)" -ForegroundColor Red
            $allHealthy = $false
        }
    }
    
    # Check TCP
    try {
        $tcp = New-Object System.Net.Sockets.TcpClient
        $tcp.Connect($TCP_HOST, $TCP_PORT)
        Write-Host "     ✅ TCP Server" -ForegroundColor Green
        $tcp.Close()
    } catch {
        Write-Host "     ❌ TCP Server" -ForegroundColor Red
        $allHealthy = $false
    }
    
    if (-not $allHealthy) {
        Write-Host ""
        Write-Host "  ⚠️  Some services are not healthy!" -ForegroundColor Yellow
        Write-Host "     Run: docker-compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d" -ForegroundColor Gray
        if (-not $Automated) {
            Write-Host ""
            $continue = Read-Host "  Continue anyway? (y/n)"
            if ($continue -ne 'y') { exit 1 }
        }
    }
    
    # Create test user
    Write-Host ""
    Write-Host "  👤 Setting up test user..." -ForegroundColor Yellow
    
    try {
        $registerBody = $TEST_USER | ConvertTo-Json
        $null = Invoke-RestMethod -Uri "$API_URL/auth/register" -Method POST -Body $registerBody -ContentType "application/json" -ErrorAction SilentlyContinue
    } catch {}
    
    try {
        $loginBody = @{ username = $TEST_USER.username; password = $TEST_USER.password } | ConvertTo-Json
        $loginResponse = Invoke-RestMethod -Uri "$API_URL/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
        $Global:AuthToken = $loginResponse.access_token
        Write-Host "     ✅ Authenticated as $($TEST_USER.username)" -ForegroundColor Green
    } catch {
        Write-Host "     ❌ Authentication failed: $($_.Exception.Message)" -ForegroundColor Red
        exit 1
    }
    
    Pause-Demo
}

# ============================================================================
# DEMO 1: DATABASE CAPACITY
# ============================================================================

function Demo-DatabaseCapacity {
    Write-Section "DATABASE CAPACITY" "30-40 manga series"
    
    $headers = @{ Authorization = "Bearer $Global:AuthToken" }
    
    Write-Host "  📖 Querying manga database..." -ForegroundColor Gray
    
    try {
        # Note: Use trailing slash to avoid redirect losing auth headers
        $response = Invoke-RestMethod -Uri "$API_URL/api/manga/?page=1&page_size=100" -Headers $headers
        $total = $response.pagination.total
        
        Write-Host ""
        Write-Host "  Database Contents:" -ForegroundColor White
        Write-Host "     └── Total manga: $total" -ForegroundColor $(if ($total -ge 30) {"Green"} else {"Yellow"})
        Write-Host ""
        
        # Show sample
        Write-Host "  Sample titles:" -ForegroundColor Gray
        $sample = $response.data | Select-Object -First 5
        foreach ($manga in $sample) {
            Write-Host "     • $($manga.title)" -ForegroundColor DarkGray
        }
        
        Write-Host ""
        Write-Result ($total -ge 30) "Manga Count" $total "≥30"
        
    } catch {
        Write-Host "  ❌ Failed to query database: $($_.Exception.Message)" -ForegroundColor Red
    }
    
    Pause-Demo
}

# ============================================================================
# DEMO 2: SEARCH LATENCY
# ============================================================================

function Demo-SearchLatency {
    Write-Section "SEARCH PERFORMANCE" "<500ms response time"
    
    $headers = @{ Authorization = "Bearer $Global:AuthToken" }
    $queries = @("naruto", "one piece", "attack", "dragon", "hero")
    
    Write-Host "  🔍 Running search queries..." -ForegroundColor Gray
    Write-Host ""
    
    $latencies = @()
    $results = @()
    
    foreach ($query in $queries) {
        $sw = [System.Diagnostics.Stopwatch]::StartNew()
        
        try {
            $encodedQuery = [System.Web.HttpUtility]::UrlEncode($query)
            $response = Invoke-RestMethod -Uri "$API_URL/api/manga/search?q=$encodedQuery" -Headers $headers -TimeoutSec 10
            $sw.Stop()
            
            $latency = $sw.ElapsedMilliseconds
            $latencies += $latency
            $count = if ($response.data) { $response.data.Count } else { 0 }
            
            $status = if ($latency -le 500) { "✅" } else { "⚠️" }
            $color = if ($latency -le 500) { "Green" } else { "Yellow" }
            
            Write-Host "     $status '$query': " -NoNewline
            Write-Host "${latency}ms" -ForegroundColor $color -NoNewline
            Write-Host " ($count results)"
            
        } catch {
            $sw.Stop()
            Write-Host "     ❌ '$query': ERROR" -ForegroundColor Red
        }
    }
    
    Write-Host ""
    
    $avgLatency = ($latencies | Measure-Object -Average).Average
    $maxLatency = ($latencies | Measure-Object -Maximum).Maximum
    $withinTarget = ($latencies | Where-Object { $_ -le 500 }).Count
    
    Write-Host "  Statistics:" -ForegroundColor White
    Write-Host "     ├── Average: $([math]::Round($avgLatency, 2))ms"
    Write-Host "     ├── Maximum: ${maxLatency}ms"
    Write-Host "     └── Within target: $withinTarget/$($queries.Count)"
    Write-Host ""
    
    Write-Result ($avgLatency -le 500) "Avg Latency" "$([math]::Round($avgLatency, 2))ms" "<500ms"
    
    Pause-Demo
}

# ============================================================================
# DEMO 3: CONCURRENT USERS
# ============================================================================

function Demo-ConcurrentUsers {
    Write-Section "CONCURRENT USERS" "50-100 simultaneous requests"
    
    $targetUsers = 100
    $headers = @{ Authorization = "Bearer $Global:AuthToken" }
    
    Write-Host "  👥 Simulating $targetUsers concurrent requests..." -ForegroundColor Gray
    Write-Host ""
    
    $jobs = @()
    $startTime = Get-Date
    
    # Create jobs
    for ($i = 0; $i -lt $targetUsers; $i++) {
        $jobs += Start-Job -ScriptBlock {
            param($url, $token)
            $headers = @{ Authorization = "Bearer $token" }
            $sw = [System.Diagnostics.Stopwatch]::StartNew()
            try {
                $response = Invoke-WebRequest -Uri "$url/api/manga?page=1&page_size=10" -Headers $headers -TimeoutSec 30 -UseBasicParsing
                $sw.Stop()
                return @{Success=$true; Latency=$sw.ElapsedMilliseconds; Status=$response.StatusCode}
            } catch {
                $sw.Stop()
                return @{Success=$false; Latency=$sw.ElapsedMilliseconds; Error=$_.Exception.Message}
            }
        } -ArgumentList $API_URL, $Global:AuthToken
        
        # Progress indicator
        if ($i % 10 -eq 0) {
            Write-Progress-Bar $i $targetUsers "Launching requests..."
        }
    }
    
    Write-Host "`r  [████████████████████████████████████████] 100% - Waiting for responses...   "
    
    # Wait and collect results
    $results = $jobs | Wait-Job | Receive-Job
    $jobs | Remove-Job
    
    $endTime = Get-Date
    $totalTime = ($endTime - $startTime).TotalSeconds
    
    # Analyze
    $successful = ($results | Where-Object { $_.Success }).Count
    $failed = $targetUsers - $successful
    $latencies = $results | Where-Object { $_.Success } | ForEach-Object { $_.Latency }
    $avgLatency = if ($latencies.Count -gt 0) { ($latencies | Measure-Object -Average).Average } else { 0 }
    $p95Index = [math]::Floor($latencies.Count * 0.95)
    $sortedLatencies = $latencies | Sort-Object
    $p95Latency = if ($sortedLatencies.Count -gt 0) { $sortedLatencies[$p95Index] } else { 0 }
    
    Write-Host ""
    Write-Host "  Results:" -ForegroundColor White
    Write-Host "     ├── Total Requests: $targetUsers"
    Write-Host "     ├── Successful: $successful ($([math]::Round($successful/$targetUsers*100, 1))%)"
    Write-Host "     ├── Failed: $failed"
    Write-Host "     ├── Avg Latency: $([math]::Round($avgLatency, 2))ms"
    Write-Host "     ├── P95 Latency: $([math]::Round($p95Latency, 2))ms"
    Write-Host "     └── Total Time: $([math]::Round($totalTime, 2))s"
    Write-Host ""
    
    $successRate = $successful / $targetUsers * 100
    Write-Result ($successRate -ge 95) "Success Rate" "$([math]::Round($successRate, 1))%" "≥95%"
    
    Pause-Demo
}

# ============================================================================
# DEMO 4: TCP CONNECTIONS
# ============================================================================

function Demo-TCPConnections {
    Write-Section "TCP CONNECTIONS" "20-30 concurrent connections"
    
    $targetConnections = 30
    
    Write-Host "  🔌 Establishing $targetConnections TCP connections..." -ForegroundColor Gray
    Write-Host ""
    
    $clients = @()
    $connected = 0
    
    for ($i = 0; $i -lt $targetConnections; $i++) {
        try {
            $client = New-Object System.Net.Sockets.TcpClient
            $client.Connect($TCP_HOST, $TCP_PORT)
            $clients += $client
            $connected++
            
            Write-Progress-Bar $connected $targetConnections "Connecting..."
        } catch {
            # Connection failed
        }
    }
    
    Write-Host "`r  [████████████████████████████████████████] 100% - Complete              "
    Write-Host ""
    
    # Verify connections are stable
    Start-Sleep -Seconds 2
    $stillConnected = ($clients | Where-Object { $_.Connected }).Count
    
    Write-Host "  Results:" -ForegroundColor White
    Write-Host "     ├── Target: $targetConnections"
    Write-Host "     ├── Connected: $connected"
    Write-Host "     └── Still Active: $stillConnected"
    Write-Host ""
    
    # Cleanup
    foreach ($client in $clients) {
        $client.Close()
    }
    
    $successRate = $connected / $targetConnections * 100
    Write-Result ($successRate -ge 80) "Connection Rate" "$([math]::Round($successRate, 1))%" "≥80%"
    
    Pause-Demo
}

# ============================================================================
# DEMO 5: MONITORING
# ============================================================================

function Demo-Monitoring {
    Write-Section "MONITORING STACK" "Prometheus + Grafana"
    
    Write-Host "  📊 Checking monitoring endpoints..." -ForegroundColor Gray
    Write-Host ""
    
    # Check Prometheus
    try {
        $promTargets = Invoke-RestMethod -Uri "$PROMETHEUS_URL/api/v1/targets" -TimeoutSec 5
        $activeTargets = ($promTargets.data.activeTargets | Where-Object { $_.health -eq "up" }).Count
        $totalTargets = $promTargets.data.activeTargets.Count
        
        Write-Host "  Prometheus:" -ForegroundColor White
        Write-Host "     ├── Status: ✅ Running" -ForegroundColor Green
        Write-Host "     └── Targets: $activeTargets/$totalTargets healthy"
    } catch {
        Write-Host "  Prometheus:" -ForegroundColor White
        Write-Host "     └── Status: ❌ Not reachable" -ForegroundColor Red
    }
    
    Write-Host ""
    
    # Check Grafana
    try {
        $grafanaHealth = Invoke-RestMethod -Uri "$GRAFANA_URL/api/health" -TimeoutSec 5
        Write-Host "  Grafana:" -ForegroundColor White
        Write-Host "     ├── Status: ✅ Running" -ForegroundColor Green
        Write-Host "     └── URL: $GRAFANA_URL (admin/mangahub_admin)"
    } catch {
        Write-Host "  Grafana:" -ForegroundColor White
        Write-Host "     └── Status: ❌ Not reachable" -ForegroundColor Red
    }
    
    Write-Host ""
    Write-Host "  📈 Quick Links:" -ForegroundColor Cyan
    Write-Host "     ├── Prometheus: $PROMETHEUS_URL/targets" -ForegroundColor Gray
    Write-Host "     ├── Grafana: $GRAFANA_URL" -ForegroundColor Gray
    Write-Host "     └── API Metrics: $API_URL/metrics" -ForegroundColor Gray
    
    if (-not $Automated) {
        Write-Host ""
        $openGrafana = Read-Host "  Open Grafana in browser? (y/n)"
        if ($openGrafana -eq 'y') {
            Start-Process $GRAFANA_URL
        }
    }
    
    Pause-Demo
}

# ============================================================================
# SUMMARY
# ============================================================================

function Show-Summary {
    Write-Banner "BENCHMARK DEMONSTRATION COMPLETE"
    
    Write-Host "  Summary of Demonstrated Capabilities:" -ForegroundColor White
    Write-Host ""
    Write-Host "     ┌─────────────────────────────┬──────────────┬────────────┐" -ForegroundColor DarkGray
    Write-Host "     │ Benchmark                   │ Target       │ Status     │" -ForegroundColor DarkGray
    Write-Host "     ├─────────────────────────────┼──────────────┼────────────┤" -ForegroundColor DarkGray
    Write-Host "     │ Concurrent Users            │ 50-100       │ ✅ Demo'd  │" -ForegroundColor White
    Write-Host "     │ Database Capacity           │ 30-40 manga  │ ✅ Demo'd  │" -ForegroundColor White
    Write-Host "     │ Search Latency              │ <500ms       │ ✅ Demo'd  │" -ForegroundColor White
    Write-Host "     │ TCP Connections             │ 20-30        │ ✅ Demo'd  │" -ForegroundColor White
    Write-Host "     │ Monitoring Stack            │ Running      │ ✅ Demo'd  │" -ForegroundColor White
    Write-Host "     └─────────────────────────────┴──────────────┴────────────┘" -ForegroundColor DarkGray
    Write-Host ""
    
    Write-Host "  📋 Next Steps:" -ForegroundColor Cyan
    Write-Host "     1. Run full benchmark suite: " -NoNewline
    Write-Host ".\run_benchmarks.ps1" -ForegroundColor Yellow
    Write-Host "     2. Run Go tests: " -NoNewline
    Write-Host "go test -v ./test/benchmark/..." -ForegroundColor Yellow
    Write-Host "     3. Run k6 load test: " -NoNewline
    Write-Host "k6 run load_test.js" -ForegroundColor Yellow
    Write-Host ""
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================

Initialize-Demo
Demo-DatabaseCapacity
Demo-SearchLatency
Demo-ConcurrentUsers
Demo-TCPConnections
Demo-Monitoring
Show-Summary
