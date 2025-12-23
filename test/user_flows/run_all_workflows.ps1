# MangaHub CLI - Master Workflow Runner
# Executes all user workflow scripts in sequence

$ErrorActionPreference = "Continue"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host "`n========================================" -ForegroundColor Magenta
Write-Host "  MangaHub CLI Master Workflow Runner" -ForegroundColor Magenta
Write-Host "========================================`n" -ForegroundColor Magenta

$workflows = @(
    @{Name="New User Registration"; Script="new_user_flow.ps1"; Description="Complete new user onboarding"},
    @{Name="Admin Operations"; Script="admin_workflow.ps1"; Description="Admin-only manga and genre management"},
    @{Name="Reader Activities"; Script="reader_workflow.ps1"; Description="Typical reader operations"},
    @{Name="TCP Sync"; Script="tcp_sync_workflow.ps1"; Description="Persistent TCP connection and sync"},
    @{Name="Multi-Protocol"; Script="multi_protocol_workflow.ps1"; Description="HTTP, gRPC, TCP, UDP communication"}
)

$results = @()

foreach ($workflow in $workflows) {
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host "  Running: $($workflow.Name)" -ForegroundColor Cyan
    Write-Host "  $($workflow.Description)" -ForegroundColor Yellow
    Write-Host "========================================`n" -ForegroundColor Cyan
    
    $scriptPath = Join-Path $ScriptDir $workflow.Script
    
    try {
        & powershell.exe -ExecutionPolicy Bypass -File $scriptPath
        
        if ($LASTEXITCODE -eq 0) {
            $results += @{Name=$workflow.Name; Status="✓ Success"; Color="Green"}
            Write-Host "`n✓ $($workflow.Name) completed successfully!" -ForegroundColor Green
        } else {
            $results += @{Name=$workflow.Name; Status="✗ Failed"; Color="Red"}
            Write-Host "`n✗ $($workflow.Name) failed with exit code $LASTEXITCODE" -ForegroundColor Red
        }
    } catch {
        $results += @{Name=$workflow.Name; Status="✗ Error"; Color="Red"}
        Write-Host "`n✗ $($workflow.Name) encountered an error: $_" -ForegroundColor Red
    }
    
    Write-Host "`nWaiting 3 seconds before next workflow..." -ForegroundColor Yellow
    Start-Sleep -Seconds 3
}

Write-Host "`n========================================" -ForegroundColor Magenta
Write-Host "  Execution Summary" -ForegroundColor Magenta
Write-Host "========================================`n" -ForegroundColor Magenta

foreach ($result in $results) {
    Write-Host "$($result.Status) - $($result.Name)" -ForegroundColor $result.Color
}

$successCount = ($results | Where-Object {$_.Status -eq "✓ Success"}).Count
$totalCount = $results.Count

Write-Host "`n$successCount/$totalCount workflows completed successfully" -ForegroundColor $(if($successCount -eq $totalCount){"Green"}else{"Yellow"})
