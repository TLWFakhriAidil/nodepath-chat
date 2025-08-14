@echo off
echo ========================================
echo Fixing Dashboard Improvements
echo ========================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [1/6] Creating backup...
mkdir fixes\backup_%date:~-4%%date:~3,2%%date:~0,2% 2>nul
copy src\views\dashboard.html fixes\backup_%date:~-4%%date:~3,2%%date:~0,2%\dashboard.html.bak >nul
copy src\ui\rest\app.go fixes\backup_%date:~-4%%date:~3,2%%date:~0,2%\app.go.bak >nul

echo [2/6] Applying dashboard fixes...
call node fixes\dashboard_improvements\apply_dashboard_fixes.js

echo [3/6] Adding worker control endpoints...
call node fixes\dashboard_improvements\add_worker_endpoints.js

echo [4/6] Fixing sequences display...
call node fixes\dashboard_improvements\fix_sequences_display.js

echo [5/6] Adding navigation to all pages...
call node fixes\dashboard_improvements\add_navigation.js

echo [6/6] Updating calendar with day labels...
call node fixes\dashboard_improvements\fix_calendar.js

echo.
echo ========================================
echo All fixes applied!
echo ========================================
echo.

REM Build and test
echo Building application...
cd src
go mod tidy
go build .
cd ..

echo.
echo Ready to commit and push changes
pause
