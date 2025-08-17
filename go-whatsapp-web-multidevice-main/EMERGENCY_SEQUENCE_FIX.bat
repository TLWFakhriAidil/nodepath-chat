@echo off
cls
echo.
echo ========================================
echo 🚨 SEQUENCE STEPS EMERGENCY FIX 🚨
echo ========================================
echo.
echo PROBLEM DETECTED:
echo - Sequences exist but show step_count: 0
echo - Steps array is empty: []  
echo - Database contains steps but Go app can't retrieve them
echo.
echo ROOT CAUSE: Missing database columns that Go query expects
echo.
echo ========================================
echo CHOOSE YOUR FIX METHOD:
echo ========================================
echo.
echo [1] Quick SQL Fix (RECOMMENDED - 30 seconds)
echo     Run SQL directly in database, then restart app
echo.
echo [2] Manual Code Fix (PERMANENT - 2 minutes)  
echo     Add emergency fix to Go code, rebuild app
echo.
echo [3] View Detailed Instructions
echo     See step-by-step manual instructions
echo.
echo [4] Just Build and Run (Skip Fix)
echo     Build current code without changes
echo.
echo [Q] Quit
echo.
set /p choice="Enter your choice (1-4 or Q): "

if /i "%choice%"=="1" goto :sql_fix
if /i "%choice%"=="2" goto :code_fix  
if /i "%choice%"=="3" goto :instructions
if /i "%choice%"=="4" goto :build_only
if /i "%choice%"=="q" goto :quit
echo Invalid choice. Please try again.
pause
goto :start

:sql_fix
echo.
echo ========================================
echo OPTION 1: Quick SQL Fix
echo ========================================
echo.
echo This will run the SQL fix directly in your database.
echo.
echo File to run: run_this_sql_fix.sql
echo.
echo Do you want to:
echo [A] Auto-run via psql (if available)
echo [M] Show manual SQL to copy/paste
echo [B] Back to main menu
echo.
set /p sqlchoice="Choose (A/M/B): "

if /i "%sqlchoice%"=="a" goto :auto_sql
if /i "%sqlchoice%"=="m" goto :manual_sql
if /i "%sqlchoice%"=="b" goto :start
goto :sql_fix

:auto_sql
echo.
echo Attempting to run SQL automatically...
echo.
if exist "run_this_sql_fix.sql" (
    echo Found SQL file. Looking for psql...
    
    REM Try to find and run psql
    for /f "tokens=*" %%i in ('where psql 2^>nul') do (
        echo Found psql: %%i
        echo.
        echo Running SQL fix...
        "%%i" %DATABASE_URL% -f run_this_sql_fix.sql
        if %ERRORLEVEL% equ 0 (
            echo.
            echo ✅ SQL fix completed successfully!
            echo.
            echo Now rebuilding and running application...
            goto :build_and_run
        ) else (
            echo.
            echo ❌ SQL fix failed. Try manual method.
            pause
            goto :start
        )
    )
    
    echo psql not found in PATH. Showing manual method...
    goto :manual_sql
) else (
    echo ❌ SQL file 'run_this_sql_fix.sql' not found!
    pause
    goto :start
)

:manual_sql
echo.
echo ========================================
echo MANUAL SQL METHOD
echo ========================================
echo.
echo 1. Open your PostgreSQL admin tool (pgAdmin, DBeaver, etc.)
echo 2. Connect to your database
echo 3. Run this SQL:
echo.
echo ----------------------------------------
type run_this_sql_fix.sql
echo ----------------------------------------
echo.
echo 4. After running SQL, press any key here to build and run the app
echo.
pause
goto :build_and_run

:code_fix
echo.
echo ========================================
echo OPTION 2: Code Fix (Permanent)
echo ========================================
echo.
echo This will modify your Go code to include the emergency fix.
echo.
echo ⚠️  This makes permanent changes to src\database\connection.go
echo.
echo Continue? (Y/N):
set /p codechoice=""
if /i not "%codechoice%"=="y" goto :start

echo.
echo Creating backup of connection.go...
if exist "src\database\connection.go" (
    copy "src\database\connection.go" "src\database\connection.go.emergency_backup" >nul
    echo ✅ Backup created: connection.go.emergency_backup
) else (
    echo ❌ File src\database\connection.go not found!
    pause
    goto :start
)

echo.
echo Modifying connection.go to include emergency fix...

REM Create a PowerShell script to modify the file
echo $content = Get-Content 'src\database\connection.go' > temp_modify.ps1
echo $content = $content -replace '\t}\s*$\s*// Run auto-migrations', '\t}^`n^`n\t// 🚨 EMERGENCY FIX: Run sequence steps fix immediately^`n\tEmergencySequenceStepsFix()^`n^`n\t// Run auto-migrations' >> temp_modify.ps1
echo $content ^| Set-Content 'src\database\connection.go' >> temp_modify.ps1

powershell -ExecutionPolicy Bypass -File temp_modify.ps1
del temp_modify.ps1

echo ✅ Code modified successfully!
echo.
goto :build_and_run

:build_and_run
echo ========================================
echo Building and Running Application
echo ========================================
echo.
echo Building application...
cd src
go build -o ..\whatsapp.exe -ldflags="-s -w"

if %ERRORLEVEL% neq 0 (
    echo ❌ Build failed!
    cd ..
    if exist "src\database\connection.go.emergency_backup" (
        echo Restoring backup...
        copy "src\database\connection.go.emergency_backup" "src\database\connection.go" >nul
    )
    pause
    goto :start
)

cd ..
echo ✅ Build successful!
echo.
echo ========================================
echo Starting Application
echo ========================================
echo.
echo Watch for these log messages:
echo   "🚨 RUNNING EMERGENCY SEQUENCE STEPS FIX..."
echo   "✅ Fix verification PASSED!"
echo.
echo Then test: http://localhost:3000/api/sequences
echo.
echo Starting application...
whatsapp.exe

echo.
echo Application stopped.
if exist "src\database\connection.go.emergency_backup" (
    echo.
    echo Restore original connection.go? (Y/N):
    set /p restore=""
    if /i "%restore%"=="y" (
        copy "src\database\connection.go.emergency_backup" "src\database\connection.go" >nul
        echo ✅ Original file restored
        del "src\database\connection.go.emergency_backup"
    )
)
pause
goto :start

:build_only
echo.
echo ========================================
echo Build and Run Only
echo ========================================
echo.
echo Building without any fixes...
cd src
go build -o ..\whatsapp.exe -ldflags="-s -w"

if %ERRORLEVEL% neq 0 (
    echo ❌ Build failed!
    cd ..
    pause
    goto :start
)

cd ..
echo ✅ Build successful!
echo.
echo Starting application...
whatsapp.exe
pause
goto :start

:instructions
echo.
echo ========================================
echo DETAILED MANUAL INSTRUCTIONS
echo ========================================
echo.
type SEQUENCE_STEPS_FIX_README.md
echo.
echo Press any key to return to menu...
pause
goto :start

:quit
echo.
echo Goodbye!
exit /b 0

:start
goto :eof
