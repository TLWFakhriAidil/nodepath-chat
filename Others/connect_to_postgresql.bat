@echo off
echo PostgreSQL Connection Helper
echo ============================
echo.

REM Check if psql is installed
where psql >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: psql is not installed or not in PATH
    echo.
    echo Please install PostgreSQL client tools:
    echo 1. Download from: https://www.postgresql.org/download/windows/
    echo 2. Or install pgAdmin: https://www.pgadmin.org/download/
    echo.
    pause
    exit /b 1
)

echo psql is installed!
echo.

REM Get connection details
echo Enter your PostgreSQL connection details:
echo.
set /p PG_HOST="Host (default: localhost): "
if "%PG_HOST%"=="" set PG_HOST=localhost

set /p PG_PORT="Port (default: 5432): "
if "%PG_PORT%"=="" set PG_PORT=5432

set /p PG_DATABASE="Database name: "
set /p PG_USER="Username: "

echo.
echo Connecting to PostgreSQL...
echo.

REM Connect to PostgreSQL
psql -h %PG_HOST% -p %PG_PORT% -U %PG_USER% -d %PG_DATABASE%

pause