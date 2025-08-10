@echo off
echo ================================================
echo WhatsApp Database Backup Tool
echo ================================================
echo.

set /p DB_HOST="Enter PostgreSQL Host (default: localhost): "
if "%DB_HOST%"=="" set DB_HOST=localhost

set /p DB_PORT="Enter PostgreSQL Port (default: 5432): "
if "%DB_PORT%"=="" set DB_PORT=5432

set /p DB_NAME="Enter Database Name: "
if "%DB_NAME%"=="" (
    echo Database name is required!
    pause
    exit /b 1
)

set /p DB_USER="Enter Database User (default: postgres): "
if "%DB_USER%"=="" set DB_USER=postgres

set /p PGPASSWORD="Enter Database Password: "

echo.
echo Choose backup type:
echo 1. Schema only (structure, no data)
echo 2. Data only (data, no structure)
echo 3. Full backup (structure + data)
echo.
set /p BACKUP_TYPE="Enter choice (1-3): "

:: Set filename with timestamp
for /f "tokens=2-4 delims=/ " %%a in ('date /t') do (set mydate=%%c-%%a-%%b)
for /f "tokens=1-2 delims=/:" %%a in ('time /t') do (set mytime=%%a%%b)
set BACKUP_FILE=whatsapp_backup_%mydate%_%mytime%.sql

echo.
echo Creating backup...

if "%BACKUP_TYPE%"=="1" (
    :: Schema only
    pg_dump -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -s -t "whatsmeow_*" > %BACKUP_FILE%
) else if "%BACKUP_TYPE%"=="2" (
    :: Data only
    pg_dump -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -a -t "whatsmeow_*" > %BACKUP_FILE%
) else if "%BACKUP_TYPE%"=="3" (
    :: Full backup
    pg_dump -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -t "whatsmeow_*" > %BACKUP_FILE%
) else (
    echo Invalid choice!
    pause
    exit /b 1
)

if %errorlevel% equ 0 (
    echo.
    echo ================================================
    echo Backup completed successfully!
    echo ================================================
    echo.
    echo Backup saved to: %BACKUP_FILE%
    echo.
    echo To restore this backup later, use:
    echo   psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% ^< %BACKUP_FILE%
    echo.
) else (
    echo.
    echo Backup failed! Check your connection details.
    echo.
)

pause
