@echo off
echo.
echo ========================================
echo Database Crash Diagnostic Tool
echo ========================================
echo.

set /p db_url="Enter your database connection string (or press Enter to use default from environment): "

if "%db_url%"=="" (
    echo Using DATABASE_URL from environment...
    set db_url=%DATABASE_URL%
)

if "%db_url%"=="" (
    echo ERROR: No database URL provided and DATABASE_URL environment variable not set!
    echo.
    echo Example format: postgresql://user:password@host:port/database
    pause
    exit /b 1
)

echo.
echo Running database diagnostics...
echo.

psql "%db_url%" -f fix_database_crash.sql

echo.
echo ========================================
echo Diagnostic complete!
echo ========================================
echo.
echo If you see columns that shouldn't exist (like 'ai', 'limit' in campaigns),
echo or tables that were added by migrations you don't want (like 'leads_ai'),
echo you can uncomment the DROP/ALTER statements in fix_database_crash.sql
echo and run this script again.
echo.
pause
