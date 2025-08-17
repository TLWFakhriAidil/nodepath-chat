@echo off
echo.
echo ========================================
echo Fixing Database Migration Crash
echo ========================================
echo.

echo Step 1: Backing up current migrations.go...
copy /Y src\database\migrations.go src\database\migrations_backup.go

echo.
echo Step 2: Applying fix - marking recent migrations as completed...
copy /Y src\database\migrations_fixed.go src\database\migrations.go

echo.
echo Step 3: Rebuilding application...
cd src
set CGO_ENABLED=0
go build -o ../whatsapp.exe .

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Build failed!
    echo Restoring original migrations.go...
    copy /Y src\database\migrations_backup.go src\database\migrations.go
    pause
    exit /b 1
)

echo.
echo ========================================
echo Fix applied successfully!
echo ========================================
echo.
echo The problematic migrations have been marked as completed.
echo The application should now start without trying to run them again.
echo.
echo If you still have issues, run diagnose_database.bat to check
echo what columns/tables exist in your database.
echo.
pause
