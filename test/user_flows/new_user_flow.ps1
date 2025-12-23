# MangaHub CLI - New User Registration and Basic Usage Flow
# This script demonstrates the complete workflow for a new user

$ErrorActionPreference = "Stop"
$CliPath = "c:\Project\Go\mangahub\cmd\cli\mangahub.exe"

Write-Host "`n========================================" -ForegroundColor Magenta
Write-Host "  New User Workflow Automation" -ForegroundColor Magenta
Write-Host "========================================`n" -ForegroundColor Magenta

# User credentials
$username = "newuser_$(Get-Date -Format 'yyyyMMddHHmmss')"
$email = "$username@mangahub.com"
$password = "Test@1234"

Write-Host "[Step 1/10] Registering new user..." -ForegroundColor Cyan
Write-Host "Username: $username" -ForegroundColor Yellow
Write-Host "Email: $email" -ForegroundColor Yellow

& $CliPath auth register --username $username --email $email --password $password
if ($LASTEXITCODE -ne 0) {
    Write-Host "Registration failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n[Step 2/10] Logging in..." -ForegroundColor Cyan
& $CliPath auth login --username $username --password $password
if ($LASTEXITCODE -ne 0) {
    Write-Host "Login failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n[Step 3/10] Browsing manga catalog..." -ForegroundColor Cyan
& $CliPath manga list --page 1 --page-size 10

Write-Host "`n[Step 4/10] Searching for manga..." -ForegroundColor Cyan
& $CliPath manga search "naruto"

Write-Host "`n[Step 5/10] Getting manga details..." -ForegroundColor Cyan
& $CliPath manga get 1

Write-Host "`n[Step 6/10] Adding manga to library..." -ForegroundColor Cyan
& $CliPath library add 1
& $CliPath library add 8
& $CliPath library add 35

Write-Host "`n[Step 7/10] Viewing library..." -ForegroundColor Cyan
& $CliPath library list

Write-Host "`n[Step 8/10] Rating manga..." -ForegroundColor Cyan
& $CliPath rating rate 1 9
& $CliPath rating rate 8 10

Write-Host "`n[Step 9/10] Adding comments..." -ForegroundColor Cyan
& $CliPath comment create 1 "This is an amazing manga! Highly recommended."
& $CliPath comment create 8 "One of the best action manga I've read!"

Write-Host "`n[Step 10/10] Updating reading progress..." -ForegroundColor Cyan
& $CliPath progress update --manga-id "1" --chapter 5 --status "reading"
& $CliPath progress update --manga-id "8" --chapter 10 --status "reading"

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "  New User Workflow Completed!" -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Green

Write-Host "User created: $username" -ForegroundColor Yellow
Write-Host "- Library: 3 manga" -ForegroundColor White
Write-Host "- Ratings: 2 ratings" -ForegroundColor White
Write-Host "- Comments: 2 comments" -ForegroundColor White
Write-Host "- Progress: 2 manga tracked" -ForegroundColor White
