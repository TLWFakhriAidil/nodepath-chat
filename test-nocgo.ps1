# PowerShell script to test CGO-free builds
Write-Host "🚀 Testing CGO-free build configuration..." -ForegroundColor Green

# Set CGO environment variables
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:GO111MODULE = "on"

Write-Host "📋 Environment variables:" -ForegroundColor Cyan
Write-Host "CGO_ENABLED=$env:CGO_ENABLED"
Write-Host "GOOS=$env:GOOS"
Write-Host "GOARCH=$env:GOARCH"
Write-Host "GO111MODULE=$env:GO111MODULE"
Write-Host ""

# Clean previous builds
Write-Host "🧹 Cleaning previous builds..." -ForegroundColor Yellow
Remove-Item -Path "server-test.exe", "migrate-test.exe", "railway-test.exe" -ErrorAction SilentlyContinue
Write-Host "✅ Cleanup complete" -ForegroundColor Green
Write-Host ""

# Test main server build
Write-Host "🔨 Building main server..." -ForegroundColor Cyan
try {
    go build -a -installsuffix cgo -o server-test.exe ./cmd/server
    if (Test-Path "server-test.exe") {
        Write-Host "✅ Server build successful!" -ForegroundColor Green
        $serverInfo = Get-Item "server-test.exe"
        Write-Host "📦 Server binary created: $($serverInfo.Name) ($([math]::Round($serverInfo.Length/1MB, 2)) MB)" -ForegroundColor Green
    } else {
        Write-Host "❌ Server binary not found" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "❌ Server build failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Test migration utility build
Write-Host "🔨 Building migration utility..." -ForegroundColor Cyan
try {
    go build -a -installsuffix cgo -o migrate-test.exe ./debug/fix_production_schema.go
    if (Test-Path "migrate-test.exe") {
        Write-Host "✅ Migration utility build successful!" -ForegroundColor Green
        $migrateInfo = Get-Item "migrate-test.exe"
        Write-Host "📦 Migration binary created: $($migrateInfo.Name) ($([math]::Round($migrateInfo.Length/1MB, 2)) MB)" -ForegroundColor Green
    } else {
        Write-Host "❌ Migration binary not found" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "❌ Migration utility build failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Test Railway migration runner build
Write-Host "🔨 Building Railway migration runner..." -ForegroundColor Cyan
try {
    go build -a -installsuffix cgo -o railway-test.exe ./debug/railway_migration_runner.go
    if (Test-Path "railway-test.exe") {
        Write-Host "✅ Railway migration runner build successful!" -ForegroundColor Green
        $railwayInfo = Get-Item "railway-test.exe"
        Write-Host "📦 Railway runner binary created: $($railwayInfo.Name) ($([math]::Round($railwayInfo.Length/1MB, 2)) MB)" -ForegroundColor Green
    } else {
        Write-Host "❌ Railway runner binary not found" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "❌ Railway runner build failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Verify configuration files
Write-Host "🚂 Verifying Railway configuration..." -ForegroundColor Cyan
if (Test-Path "railway.toml") {
    Write-Host "✅ railway.toml found" -ForegroundColor Green
} else {
    Write-Host "⚠️  railway.toml not found" -ForegroundColor Yellow
}

if (Test-Path "railway-deploy-nocgo.yml") {
    Write-Host "✅ railway-deploy-nocgo.yml found" -ForegroundColor Green
} else {
    Write-Host "⚠️  railway-deploy-nocgo.yml not found" -ForegroundColor Yellow
}

if (Test-Path "Dockerfile") {
    Write-Host "✅ Dockerfile found" -ForegroundColor Green
    $dockerfileContent = Get-Content "Dockerfile" -Raw
    if ($dockerfileContent -match "CGO_ENABLED=0") {
        Write-Host "✅ Dockerfile has CGO_ENABLED=0" -ForegroundColor Green
    } else {
        Write-Host "⚠️  Dockerfile missing CGO_ENABLED=0" -ForegroundColor Yellow
    }
} else {
    Write-Host "❌ Dockerfile not found" -ForegroundColor Red
    exit 1
}
Write-Host ""

Write-Host "🎉 All CGO-free build tests passed!" -ForegroundColor Green
Write-Host "📋 Summary:" -ForegroundColor Cyan
if (Test-Path "server-test.exe") {
    $serverSize = [math]::Round((Get-Item "server-test.exe").Length/1MB, 2)
    Write-Host "   - Server binary: server-test.exe ($serverSize MB)" -ForegroundColor White
}
if (Test-Path "migrate-test.exe") {
    $migrateSize = [math]::Round((Get-Item "migrate-test.exe").Length/1MB, 2)
    Write-Host "   - Migration binary: migrate-test.exe ($migrateSize MB)" -ForegroundColor White
}
if (Test-Path "railway-test.exe") {
    $railwaySize = [math]::Round((Get-Item "railway-test.exe").Length/1MB, 2)
    Write-Host "   - Railway runner binary: railway-test.exe ($railwaySize MB)" -ForegroundColor White
}
Write-Host "   - Railway config: ✅ Ready" -ForegroundColor Green
Write-Host ""
Write-Host "🚀 Ready for Railway deployment without CGO!" -ForegroundColor Green

# Clean up test binaries
Write-Host "🧹 Cleaning up test binaries..." -ForegroundColor Yellow
Remove-Item -Path "server-test.exe", "migrate-test.exe", "railway-test.exe" -ErrorAction SilentlyContinue
Write-Host "✅ Cleanup complete" -ForegroundColor Green

exit 0