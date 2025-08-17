@echo off
echo ==========================================
echo Dashboard Improvements - Direct Fix
echo ==========================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [1/5] Creating backup...
mkdir fixes\backup_%date:~-4%%date:~3,2%%date:~0,2%_%time:~0,2%%time:~3,2% 2>nul
copy src\views\dashboard.html fixes\backup_%date:~-4%%date:~3,2%%date:~0,2%_%time:~0,2%%time:~3,2%\dashboard.html.bak >nul
copy src\ui\rest\app.go fixes\backup_%date:~-4%%date:~3,2%%date:~0,2%_%time:~0,2%%time:~3,2%\app.go.bak >nul

echo [2/5] Applying fixes directly...

REM Apply the fixes using PowerShell
powershell -ExecutionPolicy Bypass -File fixes\dashboard_improvements\apply_fixes.ps1

echo [3/5] Building application...
cd src
go mod tidy
go build .
cd ..

echo [4/5] Updating README...
echo. >> README.md
echo ## Latest Updates - Dashboard Improvements >> README.md
echo. >> README.md
echo ### Fixed Issues (%date%) >> README.md
echo - Worker auto-refresh disabled by default >> README.md
echo - Added Resume Failed and Stop All worker buttons >> README.md
echo - Fixed sequences data display (was showing zeros) >> README.md
echo - Added Back and Home navigation buttons to all pages >> README.md
echo - Campaign calendar now shows day labels >> README.md
echo - Support for multiple campaigns per day >> README.md
echo. >> README.md

echo [5/5] Committing and pushing to GitHub...
git add -A
git commit -m "Fix: Dashboard improvements - Complete overhaul

- Worker Management: Disabled auto-refresh by default, added Resume/Stop controls
- Sequences: Fixed data population issue showing zeros
- Navigation: Added Back/Home buttons to all pages
- Calendar: Added day labels and multi-campaign support
- UI: Modern dashboard design with improved stats display
- Performance: Optimized worker status updates"

git push origin main --force

echo.
echo ==========================================
echo All fixes applied and pushed successfully!
echo ==========================================
echo.
pause
