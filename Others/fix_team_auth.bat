@echo off
echo ========================================
echo Fixing Team Member Auth Issue
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
git commit -m "Fix team member API authentication issue

- Added admin-only checks to team member management endpoints
- Fixed 401 error when accessing /api/team-members
- Added isAdminUser helper to distinguish between admin and team members
- All team member CRUD operations now require admin authentication"

git push origin main

echo.
echo ========================================
echo Fix deployed successfully!
echo ========================================
pause
