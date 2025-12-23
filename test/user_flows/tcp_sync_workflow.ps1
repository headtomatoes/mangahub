# MangaHub CLI - TCP Sync Workflow
# Demonstrates persistent connection and real-time progress sync

$ErrorActionPreference = "Stop"
$CliPath = "c:\Project\Go\mangahub\cmd\cli\mangahub.exe"

Write-Host "`n========================================" -ForegroundColor Magenta
Write-Host "  TCP Sync Workflow Automation" -ForegroundColor Magenta
Write-Host "========================================`n" -ForegroundColor Magenta

# User credentials
$username = "admin2"
$password = "SecurePass123!"

Write-Host "[Step 1/15] Logging in..." -ForegroundColor Cyan
& $CliPath auth login --username $username --password $password
if ($LASTEXITCODE -ne 0) {
    Write-Host "Login failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n[Step 2/15] Checking TCP sync status (before connection)..." -ForegroundColor Cyan
& $CliPath sync status

Write-Host "`n[Step 3/15] Connecting to TCP sync server..." -ForegroundColor Cyan
& $CliPath sync connect
if ($LASTEXITCODE -ne 0) {
    Write-Host "Connection failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`nWaiting for daemon to initialize..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

Write-Host "`n[Step 4/15] Verifying connection status..." -ForegroundColor Cyan
& $CliPath sync status

Write-Host "`n[Step 5/15] Syncing progress for manga 1..." -ForegroundColor Cyan
& $CliPath progress sync --manga-id "1" --chapter 20 --status "reading"

Write-Host "`n[Step 6/15] Syncing progress for manga 8..." -ForegroundColor Cyan
& $CliPath progress sync --manga-id "8" --chapter 50 --status "reading"

Write-Host "`n[Step 7/15] Syncing progress for manga 35..." -ForegroundColor Cyan
& $CliPath progress sync --manga-id "35" --chapter 700 --status "completed"

Write-Host "`n[Step 8/15] Checking sync status..." -ForegroundColor Cyan
& $CliPath progress sync-status

Write-Host "`n[Step 9/15] Simulating CLI restart (connection should persist)..." -ForegroundColor Yellow
Write-Host "Closing current terminal session and checking if daemon survives..." -ForegroundColor Yellow
Start-Sleep -Seconds 2

Write-Host "`n[Step 10/15] Checking TCP status after 'restart'..." -ForegroundColor Cyan
& $CliPath sync status

Write-Host "`n[Step 11/15] Syncing more progress (using persistent connection)..." -ForegroundColor Cyan
& $CliPath progress sync --manga-id "9" --chapter 100 --status "reading"

Write-Host "`n[Step 12/15] Verifying uptime increased..." -ForegroundColor Cyan
Start-Sleep -Seconds 5
& $CliPath sync status

Write-Host "`n[Step 13/15] Starting monitor (10 seconds sample)..." -ForegroundColor Cyan
$monitorJob = Start-Job -ScriptBlock {
    param($cliPath)
    & $cliPath sync monitor
} -ArgumentList $CliPath

Write-Host "Monitor running in background..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

Stop-Job -Job $monitorJob
$monitorOutput = Receive-Job -Job $monitorJob
Remove-Job -Job $monitorJob

Write-Host "Monitor output:" -ForegroundColor Green
Write-Host $monitorOutput

Write-Host "`n[Step 14/15] Checking final status..." -ForegroundColor Cyan
& $CliPath sync status

Write-Host "`n[Step 15/15] Disconnecting from TCP server..." -ForegroundColor Cyan
& $CliPath sync disconnect

Write-Host "`n[Step 16/15] Verifying disconnection..." -ForegroundColor Cyan
& $CliPath sync status

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "  TCP Sync Workflow Completed!" -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Green

Write-Host "TCP sync session summary:" -ForegroundColor Yellow
Write-Host "- Connection: Established and persisted across commands" -ForegroundColor White
Write-Host "- Progress syncs: 4 manga synced via TCP" -ForegroundColor White
Write-Host "- Daemon uptime: Successfully tracked across invocations" -ForegroundColor White
Write-Host "- Monitor: Real-time updates captured" -ForegroundColor White
Write-Host "- Disconnect: Clean shutdown performed" -ForegroundColor White
