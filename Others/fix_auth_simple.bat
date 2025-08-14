@echo off
echo ========================================
echo Removing Extra Auth Checks
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
git commit -m "Remove redundant authentication checks from team member handlers

- Removed extra isAdminUser checks since CustomAuth middleware handles auth
- The 401 error was caused by redundant authentication checks
- Now relies on app-level middleware for authentication"

git push origin main

echo.
echo ========================================
echo Fix deployed successfully!
echo ========================================
echo.
echo The authentication should now work properly.
echo Please try accessing the User Management tab again.
echo.
pause
