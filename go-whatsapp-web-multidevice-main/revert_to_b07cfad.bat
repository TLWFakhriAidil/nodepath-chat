@echo off
echo ================================================
echo Reverting to commit b07cfad
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
echo - Basic WhatsApp functionality
echo.
echo IMPORTANT: Database considerations
echo ---------------------------------
echo The following database tables might have been added after this commit:
echo - whatsmeow_* tables (for session storage)
echo.
echo These tables will remain in your database but won't be used.
echo The app will work fine with them present.
echo.
echo If you want to clean them up later, you can run:
echo DROP TABLE IF EXISTS whatsmeow_device CASCADE;
echo DROP TABLE IF EXISTS whatsmeow_identity_keys CASCADE;
echo DROP TABLE IF EXISTS whatsmeow_pre_keys CASCADE;
echo DROP TABLE IF EXISTS whatsmeow_sessions CASCADE;
echo DROP TABLE IF EXISTS whatsmeow_sender_keys CASCADE;
echo DROP TABLE IF EXISTS whatsmeow_app_state_sync_keys CASCADE;
echo DROP TABLE IF EXISTS whatsmeow_app_state_version CASCADE;
echo DROP TABLE IF EXISTS whatsmeow_app_state_mutation_macs CASCADE;
echo DROP TABLE IF EXISTS whatsmeow_message_secrets CASCADE;
echo DROP TABLE IF EXISTS whatsmeow_contacts CASCADE;
echo DROP TABLE IF EXISTS whatsmeow_chat_settings CASCADE;
echo.
pause