# PowerShell script to test database connection

# Check if dotenv is installed
$dotenvInstalled = npm list dotenv | Select-String "dotenv"
if (-not $dotenvInstalled) {
    Write-Host "Installing dotenv package..."
    npm install dotenv
}

# Run the database connection test
Write-Host "Running database connection test..."
node test-db-connection.js

# Check if the test was successful
if ($LASTEXITCODE -ne 0) {
    Write-Host "`nDatabase connection test failed with exit code $LASTEXITCODE" -ForegroundColor Red
    exit $LASTEXITCODE
} else {
    Write-Host "`nDatabase connection test completed successfully" -ForegroundColor Green
}