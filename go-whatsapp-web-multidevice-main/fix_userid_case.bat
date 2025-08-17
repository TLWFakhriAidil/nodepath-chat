@echo off
echo ========================================
echo Fixing UserID Case Sensitivity Issue
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
git commit -m "Fix UserID case sensitivity in team member authentication

- Fixed context key from 'userId' to 'UserID' (capital U)
- Added UUID parsing for user ID in CreateTeamMember
- This fixes the 401 unauthorized error when accessing team member endpoints"

git push origin main

echo.
echo ========================================
echo Fix deployed successfully!
echo ========================================
echo.
echo The authentication issue should now be resolved.
echo Try accessing the User Management tab again.
echo.
pause
