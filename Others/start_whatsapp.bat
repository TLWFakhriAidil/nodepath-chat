@echo off
echo ====================================
echo Starting WhatsApp Multi-Device
echo ====================================

REM Set DATABASE_URL from environment or use default
if "%DATABASE_URL%"=="" (
    set "DATABASE_URL=postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
)

echo Using database: %DATABASE_URL%
echo.

REM Start the application with rest API
whatsapp.exe rest --db-uri="%DATABASE_URL%"

pause
