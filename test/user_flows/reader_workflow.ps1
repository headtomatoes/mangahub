# MangaHub CLI - Reader User Workflow
# Demonstrates typical reader operations (browse, rate, comment, track progress)

$ErrorActionPreference = "Stop"
$CliPath = "c:\Project\Go\mangahub\cmd\cli\mangahub.exe"

Write-Host "`n========================================" -ForegroundColor Magenta
Write-Host "  Reader Workflow Automation" -ForegroundColor Magenta
Write-Host "========================================`n" -ForegroundColor Magenta

# Reader credentials (uses admin2 for demo, but represents a regular reader)
$username = "admin2"
$password = "SecurePass123!"

Write-Host "[Step 1/12] Logging in..." -ForegroundColor Cyan
& $CliPath auth login --username $username --password $password
if ($LASTEXITCODE -ne 0) {
    Write-Host "Login failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n[Step 2/12] Browsing new releases (latest manga)..." -ForegroundColor Cyan
& $CliPath manga list --page 1 --page-size 5

Write-Host "`n[Step 3/12] Searching for favorite manga..." -ForegroundColor Cyan
& $CliPath manga search "one piece"

Write-Host "`n[Step 4/12] Viewing manga details..." -ForegroundColor Cyan
& $CliPath manga get 9  # Jujutsu Kaisen

Write-Host "`n[Step 5/12] Adding manga to library..." -ForegroundColor Cyan
& $CliPath library add 9
& $CliPath library add 3
& $CliPath library add 25

Write-Host "`n[Step 6/12] Viewing personal library..." -ForegroundColor Cyan
& $CliPath library list

Write-Host "`n[Step 7/12] Rating manga..." -ForegroundColor Cyan
& $CliPath rating rate 9 10
& $CliPath rating rate 3 9
& $CliPath rating rate 25 8

Write-Host "`n[Step 8/12] Checking average rating..." -ForegroundColor Cyan
& $CliPath rating average 9

Write-Host "`n[Step 9/12] Viewing other readers' ratings..." -ForegroundColor Cyan
& $CliPath rating list 9 --page 1 --page-size 5

Write-Host "`n[Step 10/12] Adding comments..." -ForegroundColor Cyan
& $CliPath comment create 9 "The artwork in this manga is incredible! Chapter 272 was mind-blowing."
& $CliPath comment create 3 "Attack on Titan has one of the best plot twists in manga history."

Write-Host "`n[Step 11/12] Reading comments from other users..." -ForegroundColor Cyan
& $CliPath comment list 9 --page 1 --page-size 5

Write-Host "`n[Step 12/12] Updating reading progress..." -ForegroundColor Cyan
& $CliPath progress update --manga-id "9" --chapter 150 --status "reading"
& $CliPath progress update --manga-id "3" --chapter 141 --status "completed"
& $CliPath progress update --manga-id "25" --chapter 75 --status "reading"

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "  Reader Workflow Completed!" -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Green

Write-Host "Reading session summary:" -ForegroundColor Yellow
Write-Host "- Library: 3 manga added" -ForegroundColor White
Write-Host "- Ratings: 3 manga rated" -ForegroundColor White
Write-Host "- Comments: 2 comments posted" -ForegroundColor White
Write-Host "- Progress: 3 manga tracked (1 completed, 2 reading)" -ForegroundColor White
