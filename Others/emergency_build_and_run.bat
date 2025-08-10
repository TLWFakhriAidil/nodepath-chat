@echo off
echo.
echo ========================================
echo 🚨 EMERGENCY SEQUENCE STEPS FIX 🚨
echo ========================================
echo.
echo Problem: Sequences show step_count: 0 and empty steps array
echo Solution: Add missing database columns and fix data
echo.

REM First, let's add the emergency fix to the database connection
echo Adding emergency fix to application...

REM Create backup
copy src\database\connection.go src\database\connection.go.backup

REM Create the modified connection.go with emergency fix
echo Updating connection.go...

(
echo 	// 🚨 EMERGENCY FIX: Run sequence steps fix immediately
echo 	EmergencySequenceStepsFix()
echo.
) > temp_fix.txt

REM Insert the fix after the admin user creation
powershell -Command "(Get-Content 'src\database\connection.go') -replace '	}$([Environment]::NewLine)+	// Run auto-migrations', '	}$([Environment]::NewLine)$([Environment]::NewLine)	// 🚨 EMERGENCY FIX: Run sequence steps fix immediately$([Environment]::NewLine)	EmergencySequenceStepsFix()$([Environment]::NewLine)$([Environment]::NewLine)	// Run auto-migrations' | Set-Content 'src\database\connection.go'"

echo ✅ Emergency fix added to application

echo.
echo ========================================
echo Building application...
echo ========================================
echo.

cd src

REM Build with emergency fix included
go build -o ..\whatsapp.exe -ldflags="-s -w"

IF %ERRORLEVEL% NEQ 0 (
    echo ❌ Build failed!
    echo.
    echo Restoring backup...
    copy database\connection.go.backup database\connection.go
    pause
    exit /b 1
)

cd ..

echo ✅ Build successful with emergency fix!
echo.
echo ========================================
echo Starting application with fix...
echo ========================================
echo.
echo Watch for these log messages:
echo   "🚨 RUNNING EMERGENCY SEQUENCE STEPS FIX..."
echo   "✅ Fix verification PASSED!"
echo.
echo Then test with: http://localhost:3000/api/sequences
echo Should show step_count ^> 0 and non-empty steps array
echo.

REM Start the application
whatsapp.exe

echo.
echo Application stopped. Restoring original connection.go...
copy src\database\connection.go.backup src\database\connection.go
del src\database\connection.go.backup
del temp_fix.txt

pause
