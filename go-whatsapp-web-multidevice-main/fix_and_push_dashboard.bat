@echo off
echo ==========================================
echo Dashboard Improvements - Complete Fix
echo ==========================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [STEP 1/5] Installing Node.js dependencies...
cd fixes\dashboard_improvements
npm init -y >nul 2>&1
cd ..\..

echo [STEP 2/5] Running all fixes...
echo.

echo Applying dashboard fixes...
node fixes\dashboard_improvements\apply_dashboard_fixes.js

echo.
echo Adding worker control endpoints...
node fixes\dashboard_improvements\add_worker_endpoints.js

echo.
echo Fixing sequences display...
node fixes\dashboard_improvements\fix_sequences_display.js

echo.
echo Adding navigation to all pages...
node fixes\dashboard_improvements\add_navigation.js

echo.
echo Fixing calendar with day labels...
node fixes\dashboard_improvements\fix_calendar.js

echo.
echo [STEP 3/5] Building application...
cd src
go mod tidy
go build .
cd ..

echo.
echo [STEP 4/5] Updating README...
call :UpdateReadme

echo.
echo [STEP 5/5] Committing and pushing to GitHub...
git add -A
git commit -m "Fix: Dashboard improvements - Worker controls, Sequences display, Navigation, Calendar updates

- Disabled worker auto-refresh by default
- Added Resume Failed and Stop All worker buttons
- Fixed sequences page data population (was showing zeros)
- Added Back and Home navigation buttons to all pages
- Fixed campaign calendar with day labels
- Support for multiple campaigns per day with time display
- Added toast notifications for better user feedback
- Improved error handling and user experience"

git push origin main --force

echo.
echo ==========================================
echo All fixes applied and pushed to GitHub!
echo ==========================================
echo.
echo Summary of changes:
echo - Worker auto-refresh disabled by default
echo - Added worker control buttons (Resume/Stop)
echo - Fixed sequences data display
echo - Added navigation bar to all pages
echo - Calendar now shows day labels
echo - Multiple campaigns per day supported
echo.
pause
goto :eof

:UpdateReadme
echo ## Latest Updates - %date% >> README.md
echo. >> README.md
echo ### Dashboard Improvements >> README.md
echo - **Worker Management**: Auto-refresh disabled by default, added Resume Failed and Stop All buttons >> README.md
echo - **Sequences Display**: Fixed data population issue that was showing zeros >> README.md
echo - **Navigation**: Added Back and Home buttons to all pages for easy navigation >> README.md
echo - **Campaign Calendar**: Added day labels and support for multiple campaigns per day >> README.md
echo - **User Experience**: Added toast notifications and improved error handling >> README.md
echo. >> README.md
goto :eof
