# MangaHub CLI - Admin User Workflow
# Demonstrates admin-specific operations

$ErrorActionPreference = "Stop"
$CliPath = "c:\Project\Go\mangahub\cmd\cli\mangahub.exe"

Write-Host "`n========================================" -ForegroundColor Magenta
Write-Host "  Admin User Workflow Automation" -ForegroundColor Magenta
Write-Host "========================================`n" -ForegroundColor Magenta

# Admin credentials
$username = "admin2"
$password = "SecurePass123!"

Write-Host "[Step 1/8] Logging in as admin..." -ForegroundColor Cyan
& $CliPath auth login --username $username --password $password
if ($LASTEXITCODE -ne 0) {
    Write-Host "Login failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n[Step 2/8] Creating new manga..." -ForegroundColor Cyan
$output = & $CliPath manga create --title "Admin Test Manga $(Get-Date -Format 'yyyyMMdd-HHmmss')" --description "Created by admin workflow script" --author "Admin Author" --status "ongoing" --chapters 50 2>&1 | Out-String
Write-Host $output

# Extract manga ID from output
$mangaIdMatch = $output -match "ID:\s*(\d+)"
$newMangaId = if ($mangaIdMatch) { $Matches[1] } else { "999" }
Write-Host "Created manga ID: $newMangaId" -ForegroundColor Yellow

Write-Host "`n[Step 3/8] Updating manga details..." -ForegroundColor Cyan
& $CliPath manga update $newMangaId --description "Updated description: This manga is managed by the admin"

Write-Host "`n[Step 4/8] Creating new genre..." -ForegroundColor Cyan
& $CliPath genre create "Admin Test Genre $(Get-Date -Format 'HHmmss')"

Write-Host "`n[Step 5/8] Listing all genres..." -ForegroundColor Cyan
& $CliPath genre list

Write-Host "`n[Step 6/8] Viewing manga by genre..." -ForegroundColor Cyan
& $CliPath genre mangas 3  # Action genre

Write-Host "`n[Step 7/8] Listing all manga (paginated)..." -ForegroundColor Cyan
& $CliPath manga list --page 1 --page-size 10

Write-Host "`n[Step 8/8] Deleting test manga..." -ForegroundColor Cyan
& $CliPath manga delete $newMangaId

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "  Admin Workflow Completed!" -ForegroundColor Green
Write-Host "========================================`n" -ForegroundColor Green

Write-Host "Admin operations performed:" -ForegroundColor Yellow
Write-Host "- Created manga (ID: $newMangaId)" -ForegroundColor White
Write-Host "- Updated manga details" -ForegroundColor White
Write-Host "- Created new genre" -ForegroundColor White
Write-Host "- Deleted test manga" -ForegroundColor White
