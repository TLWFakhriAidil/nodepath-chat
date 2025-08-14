@echo off
echo ================================================
echo Quick Backup - Working Version
echo ================================================
echo.

:: Create backup directory
set BACKUP_DIR=backups\2025-07-01_00-01-03_working_version
echo Backup directory: %BACKUP_DIR%
echo.

:: Get DATABASE_URL
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
echo Creating PostgreSQL backup...
pg_dump "%DATABASE_URL%" > "%BACKUP_DIR%\postgresql_backup.sql"

if %errorlevel% equ 0 (
    echo SUCCESS: PostgreSQL backup created!
    
    :: Get current statistics
    echo. > "%BACKUP_DIR%\database_stats.txt"
    echo Database Statistics >> "%BACKUP_DIR%\database_stats.txt"
    echo ================== >> "%BACKUP_DIR%\database_stats.txt"
    echo Backup Date: %date% %time% >> "%BACKUP_DIR%\database_stats.txt"
    echo. >> "%BACKUP_DIR%\database_stats.txt"
    
    psql "%DATABASE_URL%" -t -c "SELECT 'Campaigns: ' || COUNT(*) FROM campaigns" >> "%BACKUP_DIR%\database_stats.txt" 2>nul
    psql "%DATABASE_URL%" -t -c "SELECT 'Leads: ' || COUNT(*) FROM leads" >> "%BACKUP_DIR%\database_stats.txt" 2>nul
    psql "%DATABASE_URL%" -t -c "SELECT 'Devices: ' || COUNT(*) FROM devices" >> "%BACKUP_DIR%\database_stats.txt" 2>nul
    psql "%DATABASE_URL%" -t -c "SELECT 'Messages: ' || COUNT(*) FROM broadcast_messages" >> "%BACKUP_DIR%\database_stats.txt" 2>nul
    psql "%DATABASE_URL%" -t -c "SELECT 'Sequences: ' || COUNT(*) FROM sequences" >> "%BACKUP_DIR%\database_stats.txt" 2>nul
    psql "%DATABASE_URL%" -t -c "SELECT 'Users: ' || COUNT(*) FROM users" >> "%BACKUP_DIR%\database_stats.txt" 2>nul
    
    :: Save current git commit
    echo. >> "%BACKUP_DIR%\database_stats.txt"
    echo Git Commit: >> "%BACKUP_DIR%\database_stats.txt"
    git log --oneline -1 >> "%BACKUP_DIR%\database_stats.txt" 2>nul
    
    echo.
    echo Backup completed successfully!
    echo Files created in %BACKUP_DIR%:
    echo - postgresql_backup.sql (full database)
    echo - database_stats.txt (current statistics)
    echo.
    echo File size:
    for %%A in ("%BACKUP_DIR%\postgresql_backup.sql") do echo PostgreSQL backup: %%~zA bytes
) else (
    echo ERROR: PostgreSQL backup failed!
    echo Make sure PostgreSQL client tools are installed.
)

echo.
echo ================================================
echo IMPORTANT: Keep this backup safe!
echo To restore: psql "DATABASE_URL" < postgresql_backup.sql
echo ================================================
echo.
pause
