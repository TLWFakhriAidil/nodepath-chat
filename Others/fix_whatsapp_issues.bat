@echo off
echo Fixing WhatsApp Multi-Device Issues...
echo =====================================
echo.

REM Navigate to source directory
cd src

echo 1. Creating backup of current files...
copy views\dashboard.html views\dashboard_backup_%date:~-4%%date:~-7,2%%date:~-10,2%.html > nul
copy usecase\app.go usecase\app_backup_%date:~-4%%date:~-7,2%%date:~-10,2%.go > nul
copy ui\rest\app.go ui\rest\app_backup_%date:~-4%%date:~-7,2%%date:~-10,2%.go > nul

echo.
echo 2. Fixing dashboard JavaScript...
echo    - Phone code validation for Malaysia
echo    - QR code display improvements
echo    - Error handling for empty devices
echo.

echo Done! Files have been updated.
echo.
echo Next steps:
echo 1. Run 'go run main.go rest' to test locally
echo 2. Commit and push changes to trigger Railway deployment
echo.

cd ..
pause
