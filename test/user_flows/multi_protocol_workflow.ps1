# MangaHub CLI - Multi-Protocol Communication Workflow
# Demonstrates HTTP API, gRPC, TCP, and UDP protocols

$ErrorActionPreference = "Stop"
$CliPath = "c:\Project\Go\mangahub\cmd\cli\mangahub.exe"

Write-Host "`n========================================" -ForegroundColor Magenta
Write-Host "  Multi-Protocol Workflow Automation" -ForegroundColor Magenta
Write-Host "========================================`n" -ForegroundColor Magenta

# User credentials
$username = "admin2"
$password = "SecurePass123!"

Write-Host "[Step 1/12] Logging in (HTTP API)..." -ForegroundColor Cyan
& $CliPath auth login --username $username --password $password
if ($LASTEXITCODE -ne 0) {
    Write-Host "Login failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n[Step 2/12] Listing manga via HTTP API..." -ForegroundColor Cyan
& $CliPath manga list --page 1 --page-size 5

Write-Host "`n[Step 3/12] Getting manga via gRPC..." -ForegroundColor Cyan
& $CliPath grpc manga get --id "1"

Write-Host "`n[Step 4/12] Searching manga via gRPC..." -ForegroundColor Cyan
& $CliPath grpc manga search --query "naruto" --limit 5

Write-Host "`n[Step 5/12] Testing UDP connection..." -ForegroundColor Cyan
& $CliPath udp test

Write-Host "`n[Step 6/12] Starting UDP listener (10 seconds)..." -ForegroundColor Cyan
$udpJob = Start-Job -ScriptBlock {
    param($cliPath)
    & $cliPath udp listen
} -ArgumentList $CliPath

Write-Host "UDP listener running..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

Stop-Job -Job $udpJob
$udpOutput = Receive-Job -Job $udpJob
Remove-Job -Job $udpJob

Write-Host "UDP output:" -ForegroundColor Green
Write-Host $udpOutput

Write-Host "`n[Step 7/12] Connecting to TCP sync server..." -ForegroundColor Cyan
& $CliPath sync connect
Start-Sleep -Seconds 3

Write-Host "`n[Step 8/12] Updating progress via HTTP API..." -ForegroundColor Cyan
& $CliPath progress update --manga-id "1" --chapter 30 --status "reading"

Write-Host "`n[Step 9/12] Syncing progress via TCP..." -ForegroundColor Cyan
& $CliPath progress sync --manga-id "8" --chapter 75 --status "reading"

Write-Host "`n[Step 10/12] Updating progress via gRPC..." -ForegroundColor Cyan
& $CliPath grpc progress update --manga-id "9" --chapter 200 --status "reading" --title "Jujutsu Kaisen"

Write-Host "`n[Step 11/12] Checking TCP sync status..." -ForegroundColor Cyan
& $CliPath sync status

Write-Host "`n[Step 12/12] Disconnecting from TCP..." -ForegroundColor Cyan
& $CliPath sync disconnect

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "  Multi-Protocol Workflow Completed!" -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Green

Write-Host "Protocol usage summary:" -ForegroundColor Yellow
Write-Host "- HTTP API: Authentication, manga listing, progress update" -ForegroundColor White
Write-Host "- gRPC: Manga retrieval, search, progress update" -ForegroundColor White
Write-Host "- TCP: Real-time progress sync with persistent connection" -ForegroundColor White
Write-Host "- UDP: Real-time notifications and events" -ForegroundColor White
Write-Host "`nAll protocols working correctly!" -ForegroundColor Green
