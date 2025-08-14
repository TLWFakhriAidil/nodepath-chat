@echo off
REM Run WhatsApp Multi-Device Test System

echo ========================================
echo WhatsApp Multi-Device Test Runner
echo Testing with 3000 devices (No real messages)
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\testing"

REM Check if test runner exists
if not exist test_runner.go (
    echo ERROR: test_runner.go not found!
    pause
    exit /b 1
)

REM Build the test runner
echo Building test runner...
go build -o test_runner.exe test_runner.go

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Build failed!
    pause
    exit /b 1
)

echo.
echo Build successful!
echo.

REM Set database URL if not set
if "%DATABASE_URL%"=="" (
    echo Setting default database URL...
    set DATABASE_URL=postgresql://postgres:postgres@localhost/whatsapp?sslmode=disable
)

REM Run the test
echo Starting test system...
echo.
test_runner.exe

pause
