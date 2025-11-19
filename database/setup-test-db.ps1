#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Setup test database for MangaHub integration tests

.DESCRIPTION
    Creates a separate test database and applies all migrations.
    Safe to run multiple times - will recreate the database each time.

.PARAMETER DatabaseName
    Name of the test database (default: mangahub_test)

.PARAMETER SkipMigrations
    Skip applying migrations (useful if you want to apply them manually)

.EXAMPLE
    .\setup-test-db.ps1
    
.EXAMPLE
    .\setup-test-db.ps1 -DatabaseName "mangahub_test_ci"

.EXAMPLE
    .\setup-test-db.ps1 -SkipMigrations
#>

param(
    [string]$DatabaseName = "mangahub_test",
    [switch]$SkipMigrations
)

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "╔═══════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║  MangaHub Test Database Setup                         ║" -ForegroundColor Cyan
Write-Host "╚═══════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# ============================================
# Validation
# ============================================

Write-Host "🔍 Validating environment..." -ForegroundColor Yellow

# Check if Docker is running
try {
    $null = docker ps -q 2>&1
    Write-Host "  ✅ Docker is running" -ForegroundColor Green
} catch {
    Write-Host "  ❌ Docker is not running" -ForegroundColor Red
    Write-Host "     Please start Docker Desktop first." -ForegroundColor Red
    exit 1
}

# Check if container exists and is running
$container = docker ps -q -f name=mangahub_db 2>$null
if (-not $container) {
    Write-Host "  ❌ Container 'mangahub_db' not found" -ForegroundColor Red
    Write-Host "     Run: docker compose up -d db" -ForegroundColor Yellow
    exit 1
}
Write-Host "  ✅ Database container is running" -ForegroundColor Green

# Check if migration files exist
if (-not (Test-Path "./database/migrations")) {
    Write-Host "  ❌ Migration directory not found: ./database/migrations" -ForegroundColor Red
    exit 1
}
Write-Host "  ✅ Migration files found" -ForegroundColor Green

Write-Host ""

# ============================================
# Database Creation
# ============================================

Write-Host "🗄️  Setting up database: $DatabaseName" -ForegroundColor Cyan
Write-Host ""

# Drop existing test database
Write-Host "  Dropping existing database (if exists)..." -ForegroundColor Yellow
$dropResult = docker exec mangahub_db psql -U mangahub -d postgres -c "DROP DATABASE IF EXISTS $DatabaseName;" 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✅ Dropped existing database" -ForegroundColor Green
} else {
    Write-Host "  ℹ️  No existing database to drop" -ForegroundColor Gray
}

# Create test database
Write-Host "  Creating new database..." -ForegroundColor Yellow
$createResult = docker exec mangahub_db psql -U mangahub -d postgres -c "CREATE DATABASE $DatabaseName;" 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Host "  ❌ Failed to create database" -ForegroundColor Red
    Write-Host $createResult
    exit 1
}
Write-Host "  ✅ Created database: $DatabaseName" -ForegroundColor Green

Write-Host ""

# ============================================
# Apply Migrations
# ============================================

if (-not $SkipMigrations) {
    Write-Host "🔄 Applying migrations..." -ForegroundColor Cyan
    Write-Host ""

    # Set database URL for this session
    $env:DATABASE_URL = "postgres://mangahub:mangahub_secret@localhost:5432/${DatabaseName}?sslmode=disable"

    # Check if golang-migrate is installed
    $migrateInstalled = Get-Command migrate -ErrorAction SilentlyContinue

    if ($migrateInstalled) {
        Write-Host "  Using golang-migrate CLI..." -ForegroundColor Yellow
        
        try {
            migrate -path ./database/migrations -database $env:DATABASE_URL up
            
            if ($LASTEXITCODE -eq 0) {
                Write-Host "  ✅ Migrations applied successfully" -ForegroundColor Green
            } else {
                throw "Migration failed with exit code: $LASTEXITCODE"
            }
        } catch {
            Write-Host "  ❌ Migration failed: $_" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "  Using manual SQL application..." -ForegroundColor Yellow
        Write-Host "  (Install golang-migrate for better migration management)" -ForegroundColor Gray
        Write-Host ""

        # Get migration files in order
        $migrations = Get-ChildItem ./database/migrations/*.up.sql | Sort-Object Name

        foreach ($migration in $migrations) {
            Write-Host "    Applying: $($migration.Name)..." -ForegroundColor Gray
            
            $content = Get-Content $migration.FullName -Raw
            $result = $content | docker exec -i mangahub_db psql -U mangahub -d $DatabaseName 2>&1
            
            if ($LASTEXITCODE -ne 0) {
                Write-Host "    ❌ Failed to apply: $($migration.Name)" -ForegroundColor Red
                Write-Host $result
                exit 1
            }
            
            Write-Host "    ✅ Applied: $($migration.Name)" -ForegroundColor Green
        }
    }

    Write-Host ""
}

# ============================================
# Verification
# ============================================

Write-Host "✅ Verifying database setup..." -ForegroundColor Cyan
Write-Host ""

# Check tables
Write-Host "  Checking tables..." -ForegroundColor Yellow
$tableCount = docker exec mangahub_db psql -U mangahub -d $DatabaseName -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';" 2>&1
$tableCount = $tableCount.Trim()

if ($tableCount -eq "11") {
    Write-Host "  ✅ All 11 tables created" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  Expected 11 tables, found: $tableCount" -ForegroundColor Yellow
}

# Check role column
Write-Host "  Checking users.role column..." -ForegroundColor Yellow
$roleExists = docker exec mangahub_db psql -U mangahub -d $DatabaseName -t -c "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'role';" 2>&1
$roleExists = $roleExists.Trim()

if ($roleExists -eq "1") {
    Write-Host "  ✅ users.role column exists" -ForegroundColor Green
} else {
    Write-Host "  ❌ users.role column missing (migration 003 not applied?)" -ForegroundColor Red
}

# Check indexes
Write-Host "  Checking indexes..." -ForegroundColor Yellow
$indexCount = docker exec mangahub_db psql -U mangahub -d $DatabaseName -t -c "SELECT COUNT(*) FROM pg_indexes WHERE tablename = 'refresh_tokens' AND indexname LIKE 'idx_refresh%';" 2>&1
$indexCount = $indexCount.Trim()

if ($indexCount -eq "2") {
    Write-Host "  ✅ Token indexes created" -ForegroundColor Green
} else {
    Write-Host "  ⚠️  Expected 2 token indexes, found: $indexCount" -ForegroundColor Yellow
}

Write-Host ""

# ============================================
# Summary
# ============================================

Write-Host "╔═══════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║  Test Database Ready!                                 ║" -ForegroundColor Green
Write-Host "╚═══════════════════════════════════════════════════════╝" -ForegroundColor Green
Write-Host ""
Write-Host "Database: $DatabaseName" -ForegroundColor White
Write-Host "Connection: postgres://mangahub:***@localhost:5432/$DatabaseName" -ForegroundColor White
Write-Host ""
Write-Host "To run integration tests:" -ForegroundColor Cyan
Write-Host "  `$env:DATABASE_URL = `"postgres://mangahub:mangahub_secret@localhost:5432/${DatabaseName}?sslmode=disable`"" -ForegroundColor White
Write-Host "  `$env:JWT_SECRET = `"test-secret-key-must-be-at-least-32-chars-long`"" -ForegroundColor White
Write-Host "  go test -v ./test/..." -ForegroundColor White
Write-Host ""
Write-Host "Or use Makefile:" -ForegroundColor Cyan
Write-Host "  make test-integration" -ForegroundColor White
Write-Host ""

# Set environment variables for current session
$env:DATABASE_URL = "postgres://mangahub:mangahub_secret@localhost:5432/${DatabaseName}?sslmode=disable"
Write-Host "✅ Environment variable set for this session" -ForegroundColor Green
Write-Host ""
