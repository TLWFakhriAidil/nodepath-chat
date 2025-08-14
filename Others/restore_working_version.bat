@echo off
echo ================================================
echo Restore from Working Version Backup
echo ================================================
echo.
echo WARNING: This will REPLACE your current database!
echo Make sure you want to restore from the backup.
echo.
set /p CONFIRM="Type YES to continue: "

if not "%CONFIRM%"=="YES" (
    echo Restore cancelled.
    pause
    exit /b 1
)

:: Set backup directory
set BACKUP_DIR=backups\2025-07-01_00-01-03_working_version

if not exist "%BACKUP_DIR%\postgresql_backup.sql" (
    echo ERROR: Backup file not found!
    echo Looking for: %BACKUP_DIR%\postgresql_backup.sql
    pause
    exit /b 1
)

echo.
echo Enter your Railway PostgreSQL DATABASE_URL
echo (from Railway dashboard > Postgres > Connect tab)
echo.
set /p DATABASE_URL="DATABASE_URL: "

if "%DATABASE_URL%"=="" (
    echo ERROR: DATABASE_URL is required!
    pause
    exit /b 1
)

echo.
echo Restoring database from backup...
echo This may take a few minutes...
echo.

psql "%DATABASE_URL%" < "%BACKUP_DIR%\postgresql_backup.sql"

if %errorlevel% equ 0 (
    echo.
    echo ================================================
    echo Database restored successfully!
    echo ================================================
    echo.
    echo Current database statistics:
    psql "%DATABASE_URL%" -t -c "SELECT 'Campaigns: ' || COUNT(*) FROM campaigns" 2>nul
    psql "%DATABASE_URL%" -t -c "SELECT 'Leads: ' || COUNT(*) FROM leads" 2>nul
    psql "%DATABASE_URL%" -t -c "SELECT 'Devices: ' || COUNT(*) FROM devices" 2>nul
    psql "%DATABASE_URL%" -t -c "SELECT 'Messages: ' || COUNT(*) FROM broadcast_messages" 2>nul
) else (
    echo.
    echo ERROR: Restore failed!
    echo Check the error messages above.
)

echo.
pause
