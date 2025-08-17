@echo off
echo ========================================
echo Quick Debug for Sequence Steps Issue
echo ========================================

echo.
echo Step 1: Killing any running instances...
taskkill /F /IM whatsapp.exe 2>nul
taskkill /F /IM go-whatsapp-web-multidevice.exe 2>nul

echo.
echo Step 2: Setting CGO_ENABLED=0 for build...
set CGO_ENABLED=0

echo.
echo Step 3: Building the application...
go build -o whatsapp.exe .

if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)

echo.
echo Step 4: Running with debug mode...
echo.
echo IMPORTANT: After starting, check the console logs when accessing /api/sequences
echo The logs should show:
echo   - "Getting steps for sequence: [ID]"
echo   - "Total steps in database for sequence [ID]: [count]"
echo   - "Retrieved X steps for sequence [ID]"
echo.
echo Starting server...
whatsapp.exe rest --debug=true --port=3000

pause
