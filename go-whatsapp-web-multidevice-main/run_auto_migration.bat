@echo off
echo ========================================
echo WhatsApp Multi-Device with Auto-Migration
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Please enter your PostgreSQL connection details:
echo (Press Enter to use default values)
echo.

set /p DB_HOST="PostgreSQL Host [localhost]: "
if "%DB_HOST%"=="" set DB_HOST=localhost

set /p DB_PORT="PostgreSQL Port [5432]: "
if "%DB_PORT%"=="" set DB_PORT=5432

set /p DB_USER="PostgreSQL User [postgres]: "
if "%DB_USER%"=="" set DB_USER=postgres

set /p DB_PASS="PostgreSQL Password: "
if "%DB_PASS%"=="" (
    echo ERROR: Password is required!
    pause
    exit /b 1
)

set /p DB_NAME="Database Name [whatsapp]: "
if "%DB_NAME%"=="" set DB_NAME=whatsapp

set DB_URI=postgresql://%DB_USER%:%DB_PASS%@%DB_HOST%:%DB_PORT%/%DB_NAME%

echo.
echo Using connection: postgresql://%DB_USER%:****@%DB_HOST%:%DB_PORT%/%DB_NAME%
echo.
echo The application will now:
echo 1. Connect to PostgreSQL database
echo 2. Run auto-migration from connection.go
echo 3. Fix sequence_steps table structure:
echo    - Remove: send_time, created_at, updated_at, day, schedule_time
echo    - Add: trigger, next_trigger, trigger_delay_hours, etc.
echo 4. Start REST API on port 3000
echo.
echo Starting server...
echo.

src\whatsapp.exe rest --debug=true --port=3000 --db-uri="%DB_URI%"

pause
