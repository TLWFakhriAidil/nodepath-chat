@echo off
echo ================================================
echo Complete System Backup (PostgreSQL + Redis)
echo ================================================
echo.
echo This will backup both PostgreSQL and Redis data
echo.

:: Set timestamp
set TIMESTAMP=2025-07-01_00-01-03_working_version
set BACKUP_DIR=backups\%TIMESTAMP%

echo Backup will be saved to: %BACKUP_DIR%
echo.

:: Ask for Railway DATABASE_URL
echo Step 1: PostgreSQL Backup
echo ========================
echo.
echo You need your Railway PostgreSQL connection string.
echo Find it in Railway dashboard: Postgres service > Connect tab > DATABASE_URL
echo.
set /p DATABASE_URL="Paste your Railway DATABASE_URL here: "

if "%DATABASE_URL%"=="" (
    echo DATABASE_URL is required!
    pause
    exit /b 1
)

:: Create PostgreSQL backup
echo.
echo Creating PostgreSQL backup...
echo.

:: Full database backup
pg_dump "%DATABASE_URL%" > "%BACKUP_DIR%\postgresql_full_backup.sql" 2>"%BACKUP_DIR%\pg_backup_errors.log"

if %errorlevel% equ 0 (
    echo PostgreSQL backup successful!
    
    :: Also create table list for reference
    psql "%DATABASE_URL%" -c "\dt" > "%BACKUP_DIR%\table_list.txt" 2>nul
    
    :: Get row counts
    echo. > "%BACKUP_DIR%\row_counts.txt"
    echo Table Row Counts: >> "%BACKUP_DIR%\row_counts.txt"
    echo ================= >> "%BACKUP_DIR%\row_counts.txt"
    psql "%DATABASE_URL%" -c "SELECT 'campaigns' as table_name, COUNT(*) as row_count FROM campaigns" >> "%BACKUP_DIR%\row_counts.txt" 2>nul
    psql "%DATABASE_URL%" -c "SELECT 'leads' as table_name, COUNT(*) as row_count FROM leads" >> "%BACKUP_DIR%\row_counts.txt" 2>nul
    psql "%DATABASE_URL%" -c "SELECT 'devices' as table_name, COUNT(*) as row_count FROM devices" >> "%BACKUP_DIR%\row_counts.txt" 2>nul
    psql "%DATABASE_URL%" -c "SELECT 'broadcast_messages' as table_name, COUNT(*) as row_count FROM broadcast_messages" >> "%BACKUP_DIR%\row_counts.txt" 2>nul
    psql "%DATABASE_URL%" -c "SELECT 'sequences' as table_name, COUNT(*) as row_count FROM sequences" >> "%BACKUP_DIR%\row_counts.txt" 2>nul
) else (
    echo PostgreSQL backup failed! Check pg_backup_errors.log
)

echo.
echo Step 2: Redis Backup
echo ===================
echo.

:: Ask for Redis connection
echo You need your Railway Redis connection string.
echo Find it in Railway dashboard: Redis service > Connect tab
echo.
set /p REDIS_URL="Paste your Railway REDIS_URL here (or press Enter to skip): "

if not "%REDIS_URL%"=="" (
    :: Parse Redis URL
    :: Format: redis://default:password@host:port
    for /f "tokens=3 delims=@" %%a in ("%REDIS_URL%") do set REDIS_HOST_PORT=%%a
    for /f "tokens=1 delims=:" %%a in ("%REDIS_HOST_PORT%") do set REDIS_HOST=%%a
    for /f "tokens=2 delims=:" %%a in ("%REDIS_HOST_PORT%") do set REDIS_PORT=%%a
    for /f "tokens=3 delims=:" %%a in ("%REDIS_URL%") do set REDIS_PASS_HOST=%%a
    for /f "tokens=1 delims=@" %%a in ("%REDIS_PASS_HOST%") do set REDIS_PASS=%%a
    
    echo.
    echo Connecting to Redis...
    
    :: Create Redis backup using redis-cli
    redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASS% --rdb "%BACKUP_DIR%\redis_backup.rdb" 2>"%BACKUP_DIR%\redis_backup_errors.log"
    
    if %errorlevel% equ 0 (
        echo Redis backup successful!
    ) else (
        echo Redis backup failed or redis-cli not installed.
        echo Trying alternative method...
        
        :: Alternative: Export keys and values
        echo. > "%BACKUP_DIR%\redis_keys.txt"
        redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASS% KEYS "*" >> "%BACKUP_DIR%\redis_keys.txt" 2>nul
    )
) else (
    echo Skipping Redis backup...
)

:: Create backup info file
echo. > "%BACKUP_DIR%\backup_info.txt"
echo Backup Information >> "%BACKUP_DIR%\backup_info.txt"
echo ================== >> "%BACKUP_DIR%\backup_info.txt"
echo Date: %date% %time% >> "%BACKUP_DIR%\backup_info.txt"
echo PostgreSQL URL: %DATABASE_URL% >> "%BACKUP_DIR%\backup_info.txt"
if not "%REDIS_URL%"=="" echo Redis URL: %REDIS_URL% >> "%BACKUP_DIR%\backup_info.txt"
echo. >> "%BACKUP_DIR%\backup_info.txt"
echo Git Commit: >> "%BACKUP_DIR%\backup_info.txt"
git log --oneline -1 >> "%BACKUP_DIR%\backup_info.txt" 2>nul

:: Get directory sizes
echo.
echo ================================================
echo Backup Summary
echo ================================================
echo.
echo Backup saved to: %BACKUP_DIR%
echo.
echo Files created:
dir /b "%BACKUP_DIR%"
echo.
echo To restore later:
echo - PostgreSQL: psql "DATABASE_URL" < postgresql_full_backup.sql
echo - Redis: redis-cli --rdb redis_backup.rdb
echo.
echo KEEP THIS BACKUP SAFE!
echo.
pause
