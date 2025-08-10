@echo off
echo ========================================
echo Adding Team Login Routes and README Update
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
git commit -m "Add team login routes and complete README documentation

- Added /team-login page route and handler
- Added /team-dashboard page route with authentication
- Added team member login/logout API endpoints
- Added team member info endpoint for dashboard
- Updated README with complete team member management documentation
- Added access URLs section to README
- Team members can now login at /team-login"

git push origin main

echo.
echo ========================================
echo Team Member System Complete!
echo ========================================
echo.
echo Team members can now login at:
echo https://your-domain.com/team-login
echo.
pause
