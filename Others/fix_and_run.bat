@echo off
echo ========================================
echo 🚨 EMERGENCY SEQUENCE STEPS FIX 🚨
echo ========================================
echo.
echo This will fix the sequence steps issue where:
echo - Sequences show step_count: 0
echo - Steps array is empty []
echo - But steps exist in database
echo.
echo ROOT CAUSE: Missing columns in sequence_steps table that Go application expects
echo.
pause

echo.
echo ========================================
echo Running emergency SQL fix...
echo ========================================
echo.

REM Run the SQL fix if psql is available
IF EXIST "C:\Program Files\PostgreSQL\*\bin\psql.exe" (
    echo Running SQL fix via PostgreSQL...
    FOR /F "tokens=*" %%i IN ('dir "C:\Program Files\PostgreSQL\*\bin\psql.exe" /B /S 2^>nul') DO (
        "%%i" -d %DB_URI% -f run_this_sql_fix.sql
        IF %ERRORLEVEL% EQU 0 (
            echo ✅ SQL fix applied successfully!
        ) ELSE (
            echo ❌ SQL fix failed. Please run manually in your database.
        )
        goto :continue
    )
) ELSE (
    echo ⚠️  PostgreSQL psql not found in standard location.
    echo Please run the SQL file manually in your database:
    echo.
    echo File: run_this_sql_fix.sql
    echo.
)

:continue
echo.
echo ========================================
echo Building and starting application...
echo ========================================
echo.

REM Build the application
echo Building application...
go build -o whatsapp.exe -ldflags="-s -w"

IF %ERRORLEVEL% NEQ 0 (
    echo ❌ Build failed!
    pause
    exit /b 1
)

echo ✅ Build successful!
echo.
echo Starting WhatsApp application...
echo.
echo ========================================
echo 🔍 DEBUG INFO:
echo ========================================
echo The application will show debug logs for sequence steps.
echo Look for logs like:
echo   "Getting steps for sequence: '...'"
echo   "Found X steps for sequence ..."
echo.
echo If you still see "Found 0 steps", the database 
echo connection might be using a different database.
echo ========================================
echo.

REM Start the application
whatsapp.exe

pause
