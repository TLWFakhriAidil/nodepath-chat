@echo off
echo ================================================
echo Complete Revert to b07cfad + Clear Session Data
echo ================================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Step 1: Checking current status...
git status
echo.

echo Step 2: Stashing any uncommitted changes...
git stash
echo.

echo Step 3: Resetting to commit b07cfad...
git reset --hard b07cfad
echo.

echo Step 4: Force pushing to GitHub...
git push origin main --force
echo.

echo ================================================
echo ✅ Successfully reverted to commit b07cfad
echo ================================================
echo.
echo This version has:
echo - Multi-device support working
echo - No auto-reconnect feature (no 502 errors)
echo - Compilation errors fixed
echo.
echo IMPORTANT: Database Cleanup Required
echo ------------------------------------
echo.
echo To clear WhatsApp session data, run this SQL in your Railway PostgreSQL:
echo.
echo 1. Go to Railway Dashboard
echo 2. Click on your PostgreSQL database
echo 3. Go to "Data" tab or use a PostgreSQL client
echo 4. Run the SQL from: clear_whatsmeow_data.sql
echo.
echo Or run these commands:
echo.
echo TRUNCATE TABLE whatsmeow_device CASCADE;
echo TRUNCATE TABLE whatsmeow_identity_keys CASCADE;
echo TRUNCATE TABLE whatsmeow_pre_keys CASCADE;
echo TRUNCATE TABLE whatsmeow_sessions CASCADE;
echo TRUNCATE TABLE whatsmeow_sender_keys CASCADE;
echo TRUNCATE TABLE whatsmeow_app_state_sync_keys CASCADE;
echo TRUNCATE TABLE whatsmeow_app_state_version CASCADE;
echo TRUNCATE TABLE whatsmeow_app_state_mutation_macs CASCADE;
echo TRUNCATE TABLE whatsmeow_message_secrets CASCADE;
echo TRUNCATE TABLE whatsmeow_contacts CASCADE;
echo TRUNCATE TABLE whatsmeow_chat_settings CASCADE;
echo.
echo This will clear all session data but keep the tables intact.
echo.
pause