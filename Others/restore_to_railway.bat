@echo off
echo ================================================
echo Restore to Railway PostgreSQL
echo ================================================
echo.
echo WARNING: This will restore data to your Railway database!
echo Make sure you have a recent backup before proceeding.
echo.

set /p CONFIRM="Are you sure you want to continue? (yes/no): "
if /i not "%CONFIRM%"=="yes" (
    echo Operation cancelled.
    pause
    exit /b 0
)

echo.
echo Available backup files:
dir /b *.sql 2>nul
echo.

set /p BACKUP_FILE="Enter backup filename to restore: "

if not exist "%BACKUP_FILE%" (
    echo File not found!
    pause
    exit /b 1
)

echo.
set /p DATABASE_URL="Enter your Railway DATABASE_URL: "

if "%DATABASE_URL%"=="" (
    echo DATABASE_URL is required!
    pause
    exit /b 1
)

echo.
echo Choose restore method:
echo 1. Drop and recreate WhatsApp tables (clean restore)
echo 2. Restore without dropping (may cause conflicts)
echo.
set /p RESTORE_METHOD="Enter choice (1-2): "

if "%RESTORE_METHOD%"=="1" (
    echo.
    echo Dropping existing WhatsApp tables...
    
    :: Create drop script
    echo -- Drop WhatsApp tables > drop_tables.sql
    echo DROP TABLE IF EXISTS whatsmeow_app_state_mutation_macs CASCADE; >> drop_tables.sql
    echo DROP TABLE IF EXISTS whatsmeow_app_state_sync_keys CASCADE; >> drop_tables.sql
    echo DROP TABLE IF EXISTS whatsmeow_app_state_version CASCADE; >> drop_tables.sql
    echo DROP TABLE IF EXISTS whatsmeow_chat_settings CASCADE; >> drop_tables.sql
    echo DROP TABLE IF EXISTS whatsmeow_contacts CASCADE; >> drop_tables.sql
    echo DROP TABLE IF EXISTS whatsmeow_identity_keys CASCADE; >> drop_tables.sql
    echo DROP TABLE IF EXISTS whatsmeow_message_secrets CASCADE; >> drop_tables.sql
    echo DROP TABLE IF EXISTS whatsmeow_pre_keys CASCADE; >> drop_tables.sql
    echo DROP TABLE IF EXISTS whatsmeow_privacy_tokens CASCADE; >> drop_tables.sql
    echo DROP TABLE IF EXISTS whatsmeow_sender_keys CASCADE; >> drop_tables.sql
    echo DROP TABLE IF EXISTS whatsmeow_sessions CASCADE; >> drop_tables.sql
    echo DROP TABLE IF EXISTS whatsmeow_device CASCADE; >> drop_tables.sql
    
    psql "%DATABASE_URL%" < drop_tables.sql
    del drop_tables.sql
)

echo.
echo Restoring from %BACKUP_FILE%...
psql "%DATABASE_URL%" < %BACKUP_FILE%

if %errorlevel% equ 0 (
    echo.
    echo ================================================
    echo Restore to Railway completed successfully!
    echo ================================================
    echo.
    echo Your WhatsApp tables have been restored.
    echo The application should work normally now.
    echo.
) else (
    echo.
    echo Restore failed!
    echo Check the error messages above.
    echo.
)

pause
