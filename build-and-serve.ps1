# Build and Serve Script - Matches Railway Deployment
# This script ensures your local environment matches Railway's production setup

Write-Host "🚀 Building frontend for production..." -ForegroundColor Cyan
npm run build

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Frontend build completed successfully" -ForegroundColor Green
    Write-Host "🔄 Starting Go backend server (matches Railway setup)..." -ForegroundColor Cyan
    go run cmd/server/main.go
} else {
    Write-Host "❌ Frontend build failed" -ForegroundColor Red
    exit 1
}