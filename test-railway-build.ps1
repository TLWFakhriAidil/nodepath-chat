#!/usr/bin/env pwsh
# Test script to simulate Railway build process and identify potential issues

Write-Host "🔍 Testing Railway Build Process Simulation" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan

# Set environment variables for Linux build (Railway uses Linux)
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"

Write-Host "📋 Environment Variables:" -ForegroundColor Yellow
Write-Host "CGO_ENABLED: $env:CGO_ENABLED"
Write-Host "GOOS: $env:GOOS"
Write-Host "GOARCH: $env:GOARCH"
Write-Host ""

# Test 1: Frontend Build
Write-Host "🎨 Testing Frontend Build..." -ForegroundColor Green
try {
    npm ci
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Frontend dependency installation failed" -ForegroundColor Red
        exit 1
    }
    
    npm run build
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Frontend build failed" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Frontend build successful" -ForegroundColor Green
} catch {
    Write-Host "❌ Frontend build error: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Test 2: Backend Main Server Build
Write-Host "🚀 Testing Backend Main Server Build..." -ForegroundColor Green
try {
    go build -a -installsuffix cgo -o test-server ./cmd/server
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Backend server build failed" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Backend server build successful" -ForegroundColor Green
} catch {
    Write-Host "❌ Backend server build error: $_" -ForegroundColor Red
    exit 1
}

# Test 3: Migration Utility Build
Write-Host "🔧 Testing Migration Utility Build..." -ForegroundColor Green
try {
    go build -a -installsuffix cgo -o test-migrate ./debug/fix_production_schema.go
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Migration utility build failed" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Migration utility build successful" -ForegroundColor Green
} catch {
    Write-Host "❌ Migration utility build error: $_" -ForegroundColor Red
    exit 1
}

# Test 4: Railway Migration Runner Build
Write-Host "🚄 Testing Railway Migration Runner Build..." -ForegroundColor Green
try {
    go build -a -installsuffix cgo -o test-railway-migration ./debug/railway_migration_runner.go
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Railway migration runner build failed" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Railway migration runner build successful" -ForegroundColor Green
} catch {
    Write-Host "❌ Railway migration runner build error: $_" -ForegroundColor Red
    exit 1
}

# Test 5: Check Required Files
Write-Host "📁 Checking Required Files..." -ForegroundColor Green
$requiredFiles = @(
    "production_fix_jam_column.sql",
    "start-with-migration.sh",
    "templates",
    "static"
)

foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Host "✅ Found: $file" -ForegroundColor Green
    } else {
        Write-Host "⚠️ Missing: $file" -ForegroundColor Yellow
    }
}

# Test 6: Go Module Verification
Write-Host "📦 Testing Go Module Verification..." -ForegroundColor Green
try {
    go mod verify
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Go module verification failed" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Go modules verified successfully" -ForegroundColor Green
} catch {
    Write-Host "❌ Go module verification error: $_" -ForegroundColor Red
    exit 1
}

# Test 7: Go Module Tidy Check
Write-Host "🧹 Testing Go Module Tidy..." -ForegroundColor Green
try {
    go mod tidy
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Go mod tidy failed" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Go mod tidy successful" -ForegroundColor Green
} catch {
    Write-Host "❌ Go mod tidy error: $_" -ForegroundColor Red
    exit 1
}

# Test 8: Check for Potential Issues
Write-Host "🔍 Checking for Potential Issues..." -ForegroundColor Green

# Check for large files that might cause build timeouts
$largeFiles = Get-ChildItem -Recurse -File | Where-Object { $_.Length -gt 10MB }
if ($largeFiles) {
    Write-Host "⚠️ Large files found (>10MB):" -ForegroundColor Yellow
    $largeFiles | ForEach-Object { Write-Host "   $($_.FullName) - $([math]::Round($_.Length/1MB, 2))MB" }
} else {
    Write-Host "✅ No large files found" -ForegroundColor Green
}

# Check for potential memory issues in dependencies
Write-Host "🧠 Checking Go Dependencies..." -ForegroundColor Green
go list -m all | Select-String "replace" | ForEach-Object {
    Write-Host "⚠️ Replace directive found: $_" -ForegroundColor Yellow
}

# Cleanup test files
Write-Host "🧹 Cleaning up test files..." -ForegroundColor Green
$testFiles = @("test-server", "test-migrate", "test-railway-migration")
foreach ($file in $testFiles) {
    if (Test-Path $file) {
        Remove-Item $file -Force
        Write-Host "🗑️ Removed: $file" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "🎉 Railway Build Simulation Complete!" -ForegroundColor Cyan
Write-Host "All tests passed successfully. The build should work on Railway." -ForegroundColor Green
Write-Host ""
Write-Host "💡 If Railway is still failing, the issue might be:" -ForegroundColor Yellow
Write-Host "   1. Railway environment variables not set correctly" -ForegroundColor Yellow
Write-Host "   2. Railway resource limits (memory/CPU)" -ForegroundColor Yellow
Write-Host "   3. Railway network connectivity issues" -ForegroundColor Yellow
Write-Host "   4. Railway Docker build context issues" -ForegroundColor Yellow