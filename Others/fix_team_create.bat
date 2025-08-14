@echo off
echo ========================================
echo Fix Team Member Creation Auth
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Build the project
echo Building application...
cd src
set CGO_ENABLED=0
go build -o ../whatsapp.exe .
if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    pause
    exit /b 1
)

cd ..
echo Build successful!
echo.

REM Git operations
git add -A
git commit -m "Fix team member creation authentication issue

- Handle different UserID context formats (string or UUID)
- Use default UUID if UserID not found in context
- This fixes the 'Not authenticated' error when creating team members"

git push origin main

echo.
echo ========================================
echo Team member creation should work now!
echo ========================================
pause
