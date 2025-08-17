@echo off
echo ========================================
echo Building Team Member Management System
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
echo Adding changes to git...
git add -A

echo Committing changes...
git commit -m "Add complete team member management system

- Added team_members and team_sessions tables with auto-migration
- Created TeamMember models and repository with full CRUD operations
- Added team member API handlers for create/update/delete/login
- Created User Management tab in admin dashboard
- Built team member login page at /team-login
- Created team dashboard with filtered views (devices, campaigns, sequences)
- Team members automatically see devices matching their username
- Passwords stored in plain text as requested (visible to admin)
- Team members have read-only access to their assigned devices
- Added middleware for team member authentication
- No background workers run for team members (prevents duplicates)
- Integrated with existing device filtering system"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Team Member Management System Complete!
echo ========================================
echo.
echo Features implemented:
echo - Database tables with migrations
echo - User Management tab for admins
echo - Team member login at /team-login
echo - Team dashboard with filtered data
echo - Automatic device assignment by name
echo - Read-only access for team members
echo.
pause
