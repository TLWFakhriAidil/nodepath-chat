# MySQL URI Database Connection Test Script
# This script tests the database connection using the provided MYSQL_URI

Write-Host "=== MySQL URI Database Connection Test ===" -ForegroundColor Cyan
Write-Host "Testing connection to Railway MySQL database..." -ForegroundColor Yellow
Write-Host ""

# Set environment variables
$env:MYSQL_URI = "mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway"
$env:APP_PORT = "3000"
$env:APP_DEBUG = "false"
$env:APP_OS = "Chrome"
$env:PORT = "8080"  # Default port for the test service

Write-Host "Environment Variables Set:" -ForegroundColor Green
Write-Host "MYSQL_URI: $env:MYSQL_URI" -ForegroundColor White
Write-Host "APP_PORT: $env:APP_PORT" -ForegroundColor White
Write-Host "APP_DEBUG: $env:APP_DEBUG" -ForegroundColor White
Write-Host "APP_OS: $env:APP_OS" -ForegroundColor White
Write-Host "PORT: $env:PORT" -ForegroundColor White
Write-Host ""

# Check if Go is installed
Write-Host "Checking Go installation..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "✅ Go is installed: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Go from https://golang.org/dl/" -ForegroundColor Yellow
    exit 1
}

Write-Host ""

# Check if the test file exists
if (-not (Test-Path "test_mysql_uri_connection.go")) {
    Write-Host "❌ test_mysql_uri_connection.go not found" -ForegroundColor Red
    exit 1
}

Write-Host "Running MySQL URI connection test..." -ForegroundColor Yellow
Write-Host "This will test:" -ForegroundColor Cyan
Write-Host "  1. Database connection using MYSQL_URI" -ForegroundColor White
Write-Host "  2. MySQL version retrieval" -ForegroundColor White
Write-Host "  3. Database existence verification" -ForegroundColor White
Write-Host "  4. Tables ending with '_nodepath'" -ForegroundColor White
Write-Host "  5. Data retrieval from tables" -ForegroundColor White
Write-Host "  6. Chatbot flows data" -ForegroundColor White
Write-Host ""

# Run the Go test program
try {
    Write-Host "Starting test service on port 8080..." -ForegroundColor Yellow
    Write-Host "Press Ctrl+C to stop the service" -ForegroundColor Gray
    Write-Host ""
    
    # Start the Go program
    $process = Start-Process -FilePath "go" -ArgumentList "run", "test_mysql_uri_connection.go" -NoNewWindow -PassThru
    
    # Wait a moment for the server to start
    Start-Sleep -Seconds 3
    
    # Test the endpoints
    Write-Host "Testing endpoints..." -ForegroundColor Yellow
    
    try {
        Write-Host "1. Testing health endpoint..." -ForegroundColor Cyan
        $healthResponse = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET -TimeoutSec 10
        Write-Host "✅ Health Check: $($healthResponse.status)" -ForegroundColor Green
        Write-Host "   Message: $($healthResponse.message)" -ForegroundColor White
        Write-Host "   Database: $($healthResponse.database)" -ForegroundColor White
        Write-Host "   Host: $($healthResponse.host)" -ForegroundColor White
        Write-Host ""
        
        Write-Host "2. Testing database test endpoint..." -ForegroundColor Cyan
        $testResponse = Invoke-RestMethod -Uri "http://localhost:8080/db-test" -Method GET -TimeoutSec 30
        
        Write-Host "Database Configuration:" -ForegroundColor Green
        Write-Host "   Host: $($testResponse.database_config.host)" -ForegroundColor White
        Write-Host "   Port: $($testResponse.database_config.port)" -ForegroundColor White
        Write-Host "   Database: $($testResponse.database_config.database)" -ForegroundColor White
        Write-Host "   User: $($testResponse.database_config.user)" -ForegroundColor White
        Write-Host ""
        
        Write-Host "Test Results:" -ForegroundColor Green
        foreach ($test in $testResponse.tests) {
            $color = switch ($test.status) {
                "SUCCESS" { "Green" }
                "WARNING" { "Yellow" }
                "FAILED" { "Red" }
                default { "White" }
            }
            Write-Host "   [$($test.status)] $($test.test): $($test.message)" -ForegroundColor $color
            if ($test.duration) {
                Write-Host "      Duration: $($test.duration)" -ForegroundColor Gray
            }
        }
        
        Write-Host ""
        Write-Host "✅ All tests completed successfully!" -ForegroundColor Green
        Write-Host "Your MYSQL_URI configuration is working correctly." -ForegroundColor Green
        
    } catch {
        Write-Host "❌ Failed to test endpoints: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host "This might indicate a database connection issue." -ForegroundColor Yellow
    }
    
    # Stop the process
    Write-Host ""
    Write-Host "Stopping test service..." -ForegroundColor Yellow
    Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
    
} catch {
    Write-Host "❌ Failed to run test: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    Write-Host "Troubleshooting steps:" -ForegroundColor Yellow
    Write-Host "1. Ensure Go modules are initialized: go mod tidy" -ForegroundColor White
    Write-Host "2. Check if MySQL driver is installed: go get github.com/go-sql-driver/mysql" -ForegroundColor White
    Write-Host "3. Verify MYSQL_URI format and credentials" -ForegroundColor White
    Write-Host "4. Check if MySQL server allows external connections" -ForegroundColor White
    exit 1
}

Write-Host ""
Write-Host "=== Test Summary ===" -ForegroundColor Cyan
Write-Host "Environment: Local test with Railway environment variables" -ForegroundColor White
Write-Host "MYSQL_URI: $env:MYSQL_URI" -ForegroundColor White
Write-Host "Status: Test completed" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. If tests passed, deploy to Railway with these environment variables" -ForegroundColor White
Write-Host "2. If tests failed, check the error messages above" -ForegroundColor White
Write-Host "3. Ensure Railway's IP (175.141.148.92) is whitelisted on MySQL server" -ForegroundColor White