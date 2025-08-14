@echo off
echo ================================================
echo Railway PostgreSQL Backup Tool
echo ================================================
echo.
echo This tool helps you backup Railway PostgreSQL database
echo.
echo You need your Railway PostgreSQL connection string.
echo It looks like: postgresql://postgres:password@host.railway.internal:5432/railway
echo.
echo You can find it in Railway dashboard:
echo 1. Go to your project
echo 2. Click on the Postgres service
echo 3. Go to "Connect" tab
echo 4. Copy the DATABASE_URL
echo.

set /p DATABASE_URL="Paste your Railway DATABASE_URL here: "

if "%DATABASE_URL%"=="" (
    echo DATABASE_URL is required!
    pause
    exit /b 1
)

:: Extract connection details from URL
for /f "tokens=2 delims=@" %%a in ("%DATABASE_URL%") do set HOST_PART=%%a
for /f "tokens=1 delims=/" %%a in ("%HOST_PART%") do set HOST_PORT=%%a
for /f "tokens=1 delims=:" %%a in ("%HOST_PORT%") do set PGHOST=%%a
for /f "tokens=2 delims=:" %%a in ("%HOST_PORT%") do set PGPORT=%%a

:: Set timestamp for filename
for /f "tokens=2-4 delims=/ " %%a in ('date /t') do (set mydate=%%c-%%a-%%b)
for /f "tokens=1-2 delims=/:" %%a in ('time /t') do (set mytime=%%a%%b)

echo.
echo Choose what to backup:
echo 1. WhatsApp tables only (recommended)
echo 2. Entire database
echo.
set /p BACKUP_CHOICE="Enter choice (1-2): "

set BACKUP_FILE=railway_backup_%mydate%_%mytime%.sql

echo.
echo Creating backup...

if "%BACKUP_CHOICE%"=="1" (
    :: WhatsApp tables only
    echo -- Railway WhatsApp Tables Backup > %BACKUP_FILE%
    echo -- Created: %date% %time% >> %BACKUP_FILE%
    echo -- Connection: %DATABASE_URL% >> %BACKUP_FILE%
    echo. >> %BACKUP_FILE%
    
    pg_dump "%DATABASE_URL%" -t "whatsmeow_*" >> %BACKUP_FILE%
) else (
    :: Full database
    echo -- Railway Full Database Backup > %BACKUP_FILE%
    echo -- Created: %date% %time% >> %BACKUP_FILE%
    echo. >> %BACKUP_FILE%
    
    pg_dump "%DATABASE_URL%" >> %BACKUP_FILE%
)

if %errorlevel% equ 0 (
    echo.
    echo ================================================
    echo Railway backup completed successfully!
    echo ================================================
    echo.
    echo Backup saved to: %BACKUP_FILE%
    echo File size: 
    for %%A in (%BACKUP_FILE%) do echo %%~zA bytes
    echo.
    echo IMPORTANT: Save this file safely!
    echo.
    echo To restore to Railway later:
    echo 1. Use restore_to_railway.bat
    echo 2. Or manually: psql "DATABASE_URL" ^< %BACKUP_FILE%
    echo.
) else (
    echo.
    echo Backup failed! 
    echo Make sure:
    echo 1. PostgreSQL client tools are installed
    echo 2. Your DATABASE_URL is correct
    echo 3. You can access Railway's database
    echo.
)

pause
