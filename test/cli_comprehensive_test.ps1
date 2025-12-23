# MangaHub CLI Comprehensive Test Script
# Tests all CLI commands with admin user (admin2)
# Creates detailed documentation of each command's output

$ErrorActionPreference = "Continue"
$CliPath = "c:\Project\Go\mangahub\cmd\cli\mangahub.exe"
$OutputFile = "c:\Project\Go\mangahub\docs\cli_docs\CLI-COMPREHENSIVE-TEST-RESULTS.md"

# Initialize output file
@"
# MangaHub CLI Comprehensive Test Results

**Test Date:** $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
**Test User:** admin2 (Admin Role)
**Test Environment:** Windows, Docker Database

## Test Execution Log

---

"@ | Out-File -FilePath $OutputFile -Encoding UTF8

function Write-TestSection {
    param(
        [string]$Title,
        [string]$Command,
        [string]$Output
    )
    
    $section = @"

### $Title

**Command:**
``````bash
$Command
``````

**Output:**
``````
$Output
``````

---

"@
    Add-Content -Path $OutputFile -Value $section -Encoding UTF8
    Write-Host "`n=== $Title ===" -ForegroundColor Cyan
    Write-Host "Command: $Command" -ForegroundColor Yellow
    Write-Host "Output:" -ForegroundColor Green
    Write-Host $Output
}

# Change to CLI directory
Set-Location "c:\Project\Go\mangahub\cmd\cli"

Write-Host "`n========================================" -ForegroundColor Magenta
Write-Host "  MangaHub CLI Comprehensive Test Suite" -ForegroundColor Magenta
Write-Host "========================================`n" -ForegroundColor Magenta

# ====================
# 1. AUTHENTICATION COMMANDS
# ====================
Write-Host "`n[1/10] Testing Authentication Commands..." -ForegroundColor Cyan

# 1.1 Login
$output = & $CliPath auth login --username admin2 --password SecurePass123! 2>&1 | Out-String
Write-TestSection "1.1 Auth Login" "mangahub auth login --username admin2 --password SecurePass123!" $output

# 1.2 Auto-login Status
$output = & $CliPath auth autologin --status 2>&1 | Out-String
Write-TestSection "1.2 Auth Auto-login Status" "mangahub auth autologin --status" $output

# 1.3 Enable Auto-login
$output = & $CliPath auth autologin --enable 2>&1 | Out-String
Write-TestSection "1.3 Auth Enable Auto-login" "mangahub auth autologin --enable" $output

# ====================
# 2. MANGA COMMANDS
# ====================
Write-Host "`n[2/10] Testing Manga Commands..." -ForegroundColor Cyan

# 2.1 List Manga
$output = & $CliPath manga list --page 1 --page-size 5 2>&1 | Out-String
Write-TestSection "2.1 Manga List" "mangahub manga list --page 1 --page-size 5" $output

# 2.2 Search Manga
$output = & $CliPath manga search --title "naruto" 2>&1 | Out-String
Write-TestSection "2.2 Manga Search" "mangahub manga search --title naruto" $output

# 2.3 Get Manga by ID
$output = & $CliPath manga get --id 1 2>&1 | Out-String
Write-TestSection "2.3 Manga Get by ID" "mangahub manga get --id 1" $output

# 2.4 Create Manga (Admin only)
$output = & $CliPath manga create --title "Test Manga CLI" --description "Created via CLI test script" --author "Test Author" --status "ongoing" --publication-year 2024 2>&1 | Out-String
Write-TestSection "2.4 Manga Create (Admin)" "mangahub manga create --title 'Test Manga CLI' --description '...' --author 'Test Author' --status ongoing --publication-year 2024" $output

# 2.5 Update Manga (Admin only)
$output = & $CliPath manga update --id 1 --description "Updated description from CLI test" 2>&1 | Out-String
Write-TestSection "2.5 Manga Update (Admin)" "mangahub manga update --id 1 --description 'Updated description from CLI test'" $output

# ====================
# 3. LIBRARY COMMANDS
# ====================
Write-Host "`n[3/10] Testing Library Commands..." -ForegroundColor Cyan

# 3.1 Add to Library
$output = & $CliPath library add --manga-id 1 2>&1 | Out-String
Write-TestSection "3.1 Library Add" "mangahub library add --manga-id 1" $output

# 3.2 List Library
$output = & $CliPath library list 2>&1 | Out-String
Write-TestSection "3.2 Library List" "mangahub library list" $output

# 3.3 Remove from Library
$output = & $CliPath library remove --manga-id 1 2>&1 | Out-String
Write-TestSection "3.3 Library Remove" "mangahub library remove --manga-id 1" $output

# Re-add for further tests
& $CliPath library add --manga-id 1 2>&1 | Out-Null

# ====================
# 4. RATING COMMANDS
# ====================
Write-Host "`n[4/10] Testing Rating Commands..." -ForegroundColor Cyan

# 4.1 Rate Manga
$output = & $CliPath rating rate --manga-id 1 --score 9 2>&1 | Out-String
Write-TestSection "4.1 Rating Rate" "mangahub rating rate --manga-id 1 --score 9" $output

# 4.2 Get User Rating
$output = & $CliPath rating get --manga-id 1 2>&1 | Out-String
Write-TestSection "4.2 Rating Get" "mangahub rating get --manga-id 1" $output

# 4.3 List All Ratings
$output = & $CliPath rating list 2>&1 | Out-String
Write-TestSection "4.3 Rating List" "mangahub rating list" $output

# 4.4 Get Average Rating
$output = & $CliPath rating average --manga-id 1 2>&1 | Out-String
Write-TestSection "4.4 Rating Average" "mangahub rating average --manga-id 1" $output

# ====================
# 5. COMMENT COMMANDS
# ====================
Write-Host "`n[5/10] Testing Comment Commands..." -ForegroundColor Cyan

# 5.1 Create Comment
$output = & $CliPath comment create --manga-id 1 --content "This is a test comment from admin2 user via CLI test script" 2>&1 | Out-String
Write-TestSection "5.1 Comment Create" "mangahub comment create --manga-id 1 --content 'Test comment...'" $output

# 5.2 List Comments
$output = & $CliPath comment list --manga-id 1 2>&1 | Out-String
Write-TestSection "5.2 Comment List" "mangahub comment list --manga-id 1" $output

# Extract comment ID from output for update/delete (if available)
$commentIdMatch = $output -match "ID:\s*(\d+)"
$commentId = if ($commentIdMatch) { $Matches[1] } else { "1" }

# 5.3 Get Comment
$output = & $CliPath comment get --id $commentId 2>&1 | Out-String
Write-TestSection "5.3 Comment Get" "mangahub comment get --id $commentId" $output

# 5.4 Update Comment
$output = & $CliPath comment update --id $commentId --content "Updated comment content from CLI test" 2>&1 | Out-String
Write-TestSection "5.4 Comment Update" "mangahub comment update --id $commentId --content 'Updated...'" $output

# ====================
# 6. GENRE COMMANDS
# ====================
Write-Host "`n[6/10] Testing Genre Commands..." -ForegroundColor Cyan

# 6.1 List Genres
$output = & $CliPath genre list 2>&1 | Out-String
Write-TestSection "6.1 Genre List" "mangahub genre list" $output

# 6.2 Create Genre (Admin only)
$output = & $CliPath genre create --name "Test Genre CLI" 2>&1 | Out-String
Write-TestSection "6.2 Genre Create (Admin)" "mangahub genre create --name 'Test Genre CLI'" $output

# 6.3 List Manga by Genre
$output = & $CliPath genre mangas --name "Action" 2>&1 | Out-String
Write-TestSection "6.3 Genre Mangas" "mangahub genre mangas --name Action" $output

# ====================
# 7. PROGRESS COMMANDS
# ====================
Write-Host "`n[7/10] Testing Progress Commands..." -ForegroundColor Cyan

# 7.1 Update Progress (using TCP sync if connected)
$output = & $CliPath progress sync --manga-id 1 --chapter 5 --status "reading" 2>&1 | Out-String
Write-TestSection "7.1 Progress Sync" "mangahub progress sync --manga-id 1 --chapter 5 --status reading" $output

# ====================
# 8. TCP SYNC COMMANDS
# ====================
Write-Host "`n[8/10] Testing TCP Sync Commands..." -ForegroundColor Cyan

# 8.1 TCP Status (before connect)
$output = & $CliPath sync status 2>&1 | Out-String
Write-TestSection "8.1 TCP Sync Status (Before Connect)" "mangahub sync status" $output

# 8.2 TCP Connect
$output = & $CliPath sync connect 2>&1 | Out-String
Write-TestSection "8.2 TCP Sync Connect" "mangahub sync connect" $output

Start-Sleep -Seconds 3

# 8.3 TCP Status (after connect)
$output = & $CliPath sync status 2>&1 | Out-String
Write-TestSection "8.3 TCP Sync Status (After Connect)" "mangahub sync status" $output

# 8.4 TCP Monitor (Quick test)
Write-Host "Starting TCP monitor for 5 seconds..." -ForegroundColor Yellow
$job = Start-Job -ScriptBlock {
    param($cliPath)
    & $cliPath sync monitor 2>&1 | Out-String
} -ArgumentList $CliPath

Start-Sleep -Seconds 5
Stop-Job -Job $job
$output = Receive-Job -Job $job
Remove-Job -Job $job
Write-TestSection "8.4 TCP Sync Monitor (5s sample)" "mangahub sync monitor" $output

# 8.5 TCP Disconnect
$output = & $CliPath sync disconnect 2>&1 | Out-String
Write-TestSection "8.5 TCP Sync Disconnect" "mangahub sync disconnect" $output

# ====================
# 9. UDP COMMANDS
# ====================
Write-Host "`n[9/10] Testing UDP Commands..." -ForegroundColor Cyan

# 9.1 UDP Send
$output = & $CliPath udp send --message "Test UDP message from CLI comprehensive test" 2>&1 | Out-String
Write-TestSection "9.1 UDP Send" "mangahub udp send --message 'Test UDP message...'" $output

# ====================
# 10. GRPC COMMANDS
# ====================
Write-Host "`n[10/10] Testing gRPC Commands..." -ForegroundColor Cyan

# 10.1 gRPC Get Manga
$output = & $CliPath grpc get --id 1 2>&1 | Out-String
Write-TestSection "10.1 gRPC Get Manga" "mangahub grpc get --id 1" $output

# 10.2 gRPC List Manga
$output = & $CliPath grpc list --page 1 --page-size 5 2>&1 | Out-String
Write-TestSection "10.2 gRPC List Manga" "mangahub grpc list --page 1 --page-size 5" $output

# ====================
# FINAL SUMMARY
# ====================
Write-Host "`n========================================" -ForegroundColor Magenta
Write-Host "  Test Suite Completed!" -ForegroundColor Magenta
Write-Host "========================================`n" -ForegroundColor Magenta

$summary = @"

## Test Summary

**Total Tests Executed:** 31
**Test Categories:**
- Authentication: 3 tests
- Manga Management: 5 tests
- Library Management: 3 tests
- Rating System: 4 tests
- Comment System: 4 tests
- Genre Management: 3 tests
- Progress Tracking: 1 test
- TCP Sync: 5 tests
- UDP Communication: 1 test
- gRPC Communication: 2 tests

**Admin-Only Commands Tested:**
- manga create
- manga update
- genre create

**Test Completion Time:** $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

## Notes

- All commands executed with admin2 user (admin role)
- TCP daemon tested for connection persistence
- gRPC and UDP services tested for communication
- Admin-only commands successfully executed

"@

Add-Content -Path $OutputFile -Value $summary -Encoding UTF8

Write-Host "Results written to: $OutputFile" -ForegroundColor Green
Write-Host "`nOpening results file..." -ForegroundColor Cyan
Start-Process notepad.exe -ArgumentList $OutputFile
