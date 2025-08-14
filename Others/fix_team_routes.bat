@echo off
echo ========================================
echo Fix Team Member Routes in Main App
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
git commit -m "Move team member routes to main app.go for proper authentication

- Moved team member API endpoints from init_team_member.go to app.go
- This ensures they use the same authentication middleware as other endpoints
- Fixed 401 unauthorized error by using consistent auth pattern
- Added uuid import for team member handlers"

git push origin main

echo.
echo ========================================
echo Team member routes fixed!
echo ========================================
echo.
echo The 401 error should now be resolved.
echo Try the User Management tab again.
echo.
pause
