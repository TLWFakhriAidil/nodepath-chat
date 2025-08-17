# Railway Environment Variables Setup Script
# This script sets up all required environment variables in your Railway project

Write-Host "=== Railway Environment Variables Setup ===" -ForegroundColor Cyan
Write-Host "Setting up environment variables for nodepath-chat service..." -ForegroundColor Yellow
Write-Host ""

# Check if Railway CLI is available
try {
    $railwayVersion = railway --version
    Write-Host "✅ Railway CLI is available" -ForegroundColor Green
} catch {
    Write-Host "❌ Railway CLI not found. Please install it first:" -ForegroundColor Red
    Write-Host "npm install -g @railway/cli" -ForegroundColor Yellow
    exit 1
}

# Check if logged in
try {
    $whoami = railway whoami
    Write-Host "✅ Logged in to Railway as: $whoami" -ForegroundColor Green
} catch {
    Write-Host "❌ Not logged in to Railway. Please run: railway login" -ForegroundColor Red
    exit 1
}

# Check if project is linked
try {
    $status = railway status
    Write-Host "✅ Project is linked" -ForegroundColor Green
} catch {
    Write-Host "❌ No project linked. Please run: railway link" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Setting up environment variables..." -ForegroundColor Yellow

# Define environment variables
$envVars = @{
    "MYSQL_URI" = "mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway"
    "APP_PORT" = "3000"
    "PORT" = "8080"
    "APP_DEBUG" = "false"
    "APP_OS" = "Chrome"
    "APP_ENV" = "production"
    "MAX_CONCURRENT_USERS" = "3000"
    "WEBSOCKET_ENABLED" = "true"
    "DB_MAX_OPEN_CONNS" = "100"
    "DB_MAX_IDLE_CONNS" = "10"
    "DB_CONN_MAX_LIFETIME" = "3600"
    "TZ" = "UTC"
    "LOG_LEVEL" = "info"
}

# Build the railway variables command with all --set flags
$setFlags = @()
foreach ($key in $envVars.Keys) {
    $value = $envVars[$key]
    $setFlags += "--set"
    $setFlags += "$key=$value"
    Write-Host "Preparing $key = $value" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "Setting all environment variables at once..." -ForegroundColor Yellow

try {
    # Use railway variables with multiple --set flags
    $command = "railway"
    $args = @("variables") + $setFlags
    
    Write-Host "Executing: railway variables $($setFlags -join ' ')" -ForegroundColor Gray
    & $command $args
    Write-Host "✅ All environment variables set successfully" -ForegroundColor Green
} catch {
    Write-Host "❌ Failed to set environment variables: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Trying individual variable setting..." -ForegroundColor Yellow
    
    # Fallback: set variables individually
    foreach ($key in $envVars.Keys) {
        $value = $envVars[$key]
        try {
            & railway variables --set "$key=$value"
            Write-Host "✅ $key set successfully" -ForegroundColor Green
        } catch {
            Write-Host "❌ Failed to set $key" -ForegroundColor Red
        }
    }
}

Write-Host ""
Write-Host "Verifying environment variables..." -ForegroundColor Yellow

# Get all variables to verify
try {
    Write-Host "Current environment variables:" -ForegroundColor Green
    railway variables
} catch {
    Write-Host "❌ Failed to retrieve variables" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Setup Complete ===" -ForegroundColor Cyan
Write-Host "Environment variables have been configured in Railway." -ForegroundColor Green
Write-Host ""
Write-Host "⚠️  IMPORTANT NOTICE:" -ForegroundColor Yellow
Write-Host "The database connection will still fail until you whitelist Railway's IP." -ForegroundColor Red
Write-Host "Railway's outbound IP: 175.141.148.92" -ForegroundColor White
Write-Host "MySQL Server: 159.89.198.71:3306" -ForegroundColor White
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Whitelist 175.141.148.92 on your MySQL server at 159.89.198.71" -ForegroundColor White
Write-Host "2. Deploy your application: railway up" -ForegroundColor White
Write-Host "3. Check logs: railway logs" -ForegroundColor White
Write-Host "4. Test your application endpoints" -ForegroundColor White
Write-Host ""
Write-Host "Alternative solution:" -ForegroundColor Yellow
Write-Host "Consider using Railway's managed MySQL service to avoid IP whitelist issues." -ForegroundColor White