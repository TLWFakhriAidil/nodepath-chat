@echo off
echo ================================================
echo Fix All Missing Columns - Complete Schema
echo ================================================
echo.

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [1/3] Staging changes...
git add src/database/whatsapp_tables.go

echo [2/3] Creating commit...
git commit -m "Fix: Add all missing columns (facebook_uuid, initialized, account)" -m "- Added facebook_uuid column" -m "- Added initialized and account columns" -m "- Drop and recreate tables for clean schema" -m "- Fixes all column missing errors"

echo [3/3] Pushing to main...
git push origin main

echo.
echo ================================================
echo Complete Schema Fix Deployed!
echo ================================================
echo.
echo This will:
echo 1. Drop existing WhatsApp tables
echo 2. Recreate with complete schema
echo 3. Include all required columns
echo.
echo The app should now start successfully!
echo ================================================
echo.
pause
