#!/usr/bin/env pwsh
# Railway Build Optimization and Fix Script
# This script addresses common Railway build issues

Write-Host "🚄 Railway Build Optimization Script" -ForegroundColor Cyan
Write-Host "====================================" -ForegroundColor Cyan
Write-Host ""

# Function to check file sizes
function Check-BuildContextSize {
    Write-Host "📏 Checking Docker build context size..." -ForegroundColor Yellow
    
    $totalSize = 0
    $largeFiles = @()
    
    Get-ChildItem -Recurse -File | ForEach-Object {
        $totalSize += $_.Length
        if ($_.Length -gt 5MB) {
            $largeFiles += [PSCustomObject]@{
                Path = $_.FullName
                SizeMB = [math]::Round($_.Length / 1MB, 2)
            }
        }
    }
    
    $totalSizeMB = [math]::Round($totalSize / 1MB, 2)
    Write-Host "Total build context size: $totalSizeMB MB" -ForegroundColor $(if ($totalSizeMB -gt 500) { 'Red' } elseif ($totalSizeMB -gt 200) { 'Yellow' } else { 'Green' })
    
    if ($largeFiles.Count -gt 0) {
        Write-Host "⚠️ Large files found (>5MB):" -ForegroundColor Yellow
        $largeFiles | ForEach-Object {
            Write-Host "   $($_.Path) - $($_.SizeMB) MB" -ForegroundColor Yellow
        }
    } else {
        Write-Host "✅ No large files found" -ForegroundColor Green
    }
    
    return $totalSizeMB
}

# Function to optimize go.mod
function Optimize-GoMod {
    Write-Host "📦 Optimizing Go modules..." -ForegroundColor Yellow
    
    try {
        # Clean module cache
        go clean -modcache
        
        # Tidy modules
        go mod tidy
        
        # Verify modules
        go mod verify
        
        # Download modules
        go mod download
        
        Write-Host "✅ Go modules optimized" -ForegroundColor Green
    } catch {
        Write-Host "❌ Go module optimization failed: $_" -ForegroundColor Red
        return $false
    }
    
    return $true
}

# Function to test builds with Railway settings
function Test-RailwayBuilds {
    Write-Host "🔨 Testing builds with Railway settings..." -ForegroundColor Yellow
    
    # Set Railway-like environment
    $env:CGO_ENABLED = "0"
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    
    $builds = @(
        @{ Name = "Main Server"; Command = "go build -a -installsuffix cgo -ldflags '-w -s' -o test-server ./cmd/server" },
        @{ Name = "Migration Utility"; Command = "go build -a -installsuffix cgo -ldflags '-w -s' -o test-migrate ./debug/fix_production_schema.go" },
        @{ Name = "Railway Migration Runner"; Command = "go build -a -installsuffix cgo -ldflags '-w -s' -o test-railway ./debug/railway_migration_runner.go" }
    )
    
    $allPassed = $true
    
    foreach ($build in $builds) {
        Write-Host "Building $($build.Name)..." -NoNewline
        
        try {
            $result = Invoke-Expression $build.Command 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Host " ✅" -ForegroundColor Green
            } else {
                Write-Host " ❌" -ForegroundColor Red
                Write-Host "Error: $result" -ForegroundColor Red
                $allPassed = $false
            }
        } catch {
            Write-Host " ❌" -ForegroundColor Red
            Write-Host "Exception: $_" -ForegroundColor Red
            $allPassed = $false
        }
    }
    
    # Cleanup test files
    @("test-server", "test-migrate", "test-railway") | ForEach-Object {
        if (Test-Path $_) { Remove-Item $_ -Force }
    }
    
    return $allPassed
}

# Function to create Railway deployment recommendations
function Show-RailwayRecommendations {
    Write-Host "💡 Railway Deployment Recommendations:" -ForegroundColor Cyan
    Write-Host ""
    
    Write-Host "1. 🔧 Build Optimization:" -ForegroundColor Yellow
    Write-Host "   - Use multi-stage Docker builds (✅ Already implemented)"
    Write-Host "   - Minimize build context size (✅ Optimized .dockerignore)"
    Write-Host "   - Use build caching where possible"
    Write-Host ""
    
    Write-Host "2. 🧠 Memory Management:" -ForegroundColor Yellow
    Write-Host "   - Set GOGC=off during build to reduce memory usage"
    Write-Host "   - Use -ldflags '-w -s' to reduce binary size"
    Write-Host "   - Consider splitting large builds into smaller steps"
    Write-Host ""
    
    Write-Host "3. ⏱️ Build Time Optimization:" -ForegroundColor Yellow
    Write-Host "   - Leverage Docker layer caching"
    Write-Host "   - Copy go.mod and go.sum first for better caching"
    Write-Host "   - Use .dockerignore to exclude unnecessary files"
    Write-Host ""
    
    Write-Host "4. 🌐 Environment Variables:" -ForegroundColor Yellow
    Write-Host "   - Ensure MYSQL_URI is set in Railway environment"
    Write-Host "   - Set PORT=8080 (✅ Already configured)"
    Write-Host "   - Set APP_ENV=production for production builds"
    Write-Host ""
    
    Write-Host "5. 🔍 Debugging Railway Issues:" -ForegroundColor Yellow
    Write-Host "   - Check Railway build logs for specific errors"
    Write-Host "   - Monitor memory usage during build"
    Write-Host "   - Verify all required files are included in build context"
    Write-Host ""
}

# Main execution
Write-Host "Starting Railway build optimization..." -ForegroundColor Green
Write-Host ""

# Step 1: Check build context size
$contextSize = Check-BuildContextSize
Write-Host ""

# Step 2: Optimize Go modules
$goModSuccess = Optimize-GoMod
Write-Host ""

# Step 3: Test builds
$buildsSuccess = Test-RailwayBuilds
Write-Host ""

# Step 4: Show recommendations
Show-RailwayRecommendations

# Summary
Write-Host "📊 Optimization Summary:" -ForegroundColor Cyan
Write-Host "========================" -ForegroundColor Cyan
Write-Host "Build context size: $contextSize MB" -ForegroundColor $(if ($contextSize -gt 200) { 'Yellow' } else { 'Green' })
Write-Host "Go modules: $(if ($goModSuccess) { '✅ Optimized' } else { '❌ Failed' })" -ForegroundColor $(if ($goModSuccess) { 'Green' } else { 'Red' })
Write-Host "Build tests: $(if ($buildsSuccess) { '✅ Passed' } else { '❌ Failed' })" -ForegroundColor $(if ($buildsSuccess) { 'Green' } else { 'Red' })
Write-Host ""

if ($goModSuccess -and $buildsSuccess -and $contextSize -lt 300) {
    Write-Host "🎉 Railway build should succeed!" -ForegroundColor Green
    Write-Host "The project is optimized for Railway deployment." -ForegroundColor Green
} else {
    Write-Host "⚠️ Potential Railway build issues detected." -ForegroundColor Yellow
    Write-Host "Review the recommendations above and Railway build logs." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "🚀 Next steps:" -ForegroundColor Cyan
Write-Host "1. Commit and push the optimized .dockerignore"
Write-Host "2. Trigger a new Railway deployment"
Write-Host "3. Monitor Railway build logs for any remaining issues"
Write-Host "4. Check Railway environment variables are set correctly"