@echo off
echo ================================================
echo Fix Missing 'lid' Column Error
echo ================================================
echo.

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [1/3] Staging changes...
git add src/database/whatsapp_tables.go

echo [2/3] Creating commit...
git commit -m "Fix: Add missing 'lid' column to whatsmeow_device table" -m "- Added lid column to device table schema" -m "- Also runs ALTER TABLE to add column to existing tables" -m "- Fixes 'column lid does not exist' error"

echo [3/3] Pushing to main...
git push origin main

echo.
echo ================================================
echo Fix Deployed!
echo ================================================
echo.
echo The 'lid' column will be added to the device table.
echo Railway should redeploy and the error will be fixed.
echo.
echo ================================================
echo.
pause
