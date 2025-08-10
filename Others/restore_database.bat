@echo off
echo ================================================
echo WhatsApp Database Restore Tool
echo ================================================
echo.

echo Available restore options:
echo 1. Restore from backup file
echo 2. Restore WhatsApp tables only (using backup_whatsapp_schema.sql)
echo 3. Quick fix - Add missing columns to existing tables
echo.
set /p RESTORE_TYPE="Enter choice (1-3): "

if "%RESTORE_TYPE%"=="1" (
    :: Restore from backup file
    echo.
    echo Available backup files:
    dir /b *.sql
    echo.
    set /p BACKUP_FILE="Enter backup filename: "
    
    if not exist "%BACKUP_FILE%" (
        echo File not found!
        pause
        exit /b 1
    )
    
    set /p DB_HOST="Enter PostgreSQL Host (default: localhost): "
    if "%DB_HOST%"=="" set DB_HOST=localhost
    
    set /p DB_PORT="Enter PostgreSQL Port (default: 5432): "
    if "%DB_PORT%"=="" set DB_PORT=5432
    
    set /p DB_NAME="Enter Database Name: "
    set /p DB_USER="Enter Database User (default: postgres): "
    if "%DB_USER%"=="" set DB_USER=postgres
    
    set /p PGPASSWORD="Enter Database Password: "
    
    echo.
    echo Restoring from %BACKUP_FILE%...
    psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% < %BACKUP_FILE%
    
) else if "%RESTORE_TYPE%"=="2" (
    :: Restore WhatsApp tables only
    set /p DB_HOST="Enter PostgreSQL Host (default: localhost): "
    if "%DB_HOST%"=="" set DB_HOST=localhost
    
    set /p DB_PORT="Enter PostgreSQL Port (default: 5432): "
    if "%DB_PORT%"=="" set DB_PORT=5432
    
    set /p DB_NAME="Enter Database Name: "
    set /p DB_USER="Enter Database User (default: postgres): "
    if "%DB_USER%"=="" set DB_USER=postgres
    
    set /p PGPASSWORD="Enter Database Password: "
    
    echo.
    echo Restoring WhatsApp tables...
    psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% < backup_whatsapp_schema.sql
    
) else if "%RESTORE_TYPE%"=="3" (
    :: Quick fix - add missing columns
    set /p DB_HOST="Enter PostgreSQL Host (default: localhost): "
    if "%DB_HOST%"=="" set DB_HOST=localhost
    
    set /p DB_PORT="Enter PostgreSQL Port (default: 5432): "
    if "%DB_PORT%"=="" set DB_PORT=5432
    
    set /p DB_NAME="Enter Database Name: "
    set /p DB_USER="Enter Database User (default: postgres): "
    if "%DB_USER%"=="" set DB_USER=postgres
    
    set /p PGPASSWORD="Enter Database Password: "
    
    echo.
    echo Adding missing columns...
    
    echo ALTER TABLE whatsmeow_device ADD COLUMN IF NOT EXISTS lid TEXT; > quick_fix.sql
    echo ALTER TABLE whatsmeow_device ADD COLUMN IF NOT EXISTS facebook_uuid TEXT; >> quick_fix.sql
    echo ALTER TABLE whatsmeow_device ADD COLUMN IF NOT EXISTS initialized BOOLEAN DEFAULT false; >> quick_fix.sql
    echo ALTER TABLE whatsmeow_device ADD COLUMN IF NOT EXISTS account BYTEA; >> quick_fix.sql
    
    psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% < quick_fix.sql
    del quick_fix.sql
    
) else (
    echo Invalid choice!
    pause
    exit /b 1
)

if %errorlevel% equ 0 (
    echo.
    echo ================================================
    echo Restore completed successfully!
    echo ================================================
    echo.
) else (
    echo.
    echo Restore failed! Check your connection details.
    echo.
)

pause
