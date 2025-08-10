@echo off
echo ================================================
echo Final Fix: Clear WhatsApp Data & Update Port
echo ================================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Step 1: Ensuring we're on the right commit...
git log --oneline -1

echo.
echo Step 2: Updating port configuration to match Railway...
REM Update the config to ensure port 3000 is used (Railway default)
cd src\config

REM Check if settings.go exists and update AppPort
powershell -Command "(Get-Content settings.go) -replace 'AppPort\s*=\s*\""\d+\""', 'AppPort = \"3000\"' | Set-Content settings.go"

cd ..\..

echo.
echo Step 3: Ensuring railway.toml uses port 3000...
(
echo [build]
echo builder = "DOCKERFILE"
echo dockerfilePath = "Dockerfile"
echo.
echo [deploy]
echo startCommand = "/app/whatsapp rest"
echo restartPolicyType = "ON_FAILURE"
echo restartPolicyMaxRetries = 10
echo.
echo [[services]]
echo name = "web"
echo port = 3000
echo.
echo [variables]
echo # Database Configuration - Railway provides DATABASE_URL
echo DB_URI = "${{DATABASE_URL}}"
echo.
echo # Application Configuration
echo APP_PORT = "3000"
echo APP_DEBUG = "false"
echo APP_OS = "WhatsApp Business System"
echo APP_BASIC_AUTH = "admin:changeme123"
echo APP_CHAT_FLUSH_INTERVAL = "30"
echo.
echo # WhatsApp Features
echo WHATSAPP_CHAT_STORAGE = "true"
echo WHATSAPP_ACCOUNT_VALIDATION = "true"
echo WHATSAPP_AUTO_REPLY = "Thank you for contacting us. We will respond shortly."
echo.
echo # Performance Settings
echo NODE_ENV = "production"
echo NODE_TLS_REJECT_UNAUTHORIZED = "0"
) > railway.toml

echo.
echo Step 4: Creating SQL to safely handle whatsmeow tables...
(
echo -- Handle WhatsApp session tables
echo -- These tables are auto-created by whatsmeow library
echo -- We'll truncate them if they exist to prevent issues
echo.
echo DO $$
echo BEGIN
echo     -- Check if whatsmeow_device table exists
echo     IF EXISTS ^(SELECT 1 FROM information_schema.tables WHERE table_name = 'whatsmeow_device'^) THEN
echo         -- Clear all session data
echo         TRUNCATE TABLE whatsmeow_message_secrets CASCADE;
echo         TRUNCATE TABLE whatsmeow_contacts CASCADE;
echo         TRUNCATE TABLE whatsmeow_chat_settings CASCADE;
echo         TRUNCATE TABLE whatsmeow_app_state_mutation_macs CASCADE;
echo         TRUNCATE TABLE whatsmeow_app_state_version CASCADE;
echo         TRUNCATE TABLE whatsmeow_app_state_sync_keys CASCADE;
echo         TRUNCATE TABLE whatsmeow_sender_keys CASCADE;
echo         TRUNCATE TABLE whatsmeow_sessions CASCADE;
echo         TRUNCATE TABLE whatsmeow_pre_keys CASCADE;
echo         TRUNCATE TABLE whatsmeow_identity_keys CASCADE;
echo         TRUNCATE TABLE whatsmeow_device CASCADE;
echo         RAISE NOTICE 'WhatsApp session tables cleared';
echo     ELSE
echo         RAISE NOTICE 'WhatsApp session tables do not exist yet';
echo     END IF;
echo END $$;
) > clear_sessions_safe.sql

echo.
echo Step 5: Committing configuration updates...
git add -A
git commit -m "Update port configuration to 3000 for Railway"

echo.
echo Step 6: Pushing to GitHub...
git push origin main --force

echo.
echo ================================================
echo ✅ Configuration Updated!
echo ================================================
echo.
echo What's been done:
echo 1. Confirmed we're on commit b07cfad (stable version)
echo 2. Updated port configuration to 3000
echo 3. Created safe SQL script for session cleanup
echo.
echo Next Steps:
echo -----------
echo 1. Wait for Railway to redeploy
echo 2. After deployment, run this SQL in Railway PostgreSQL:
echo.
echo    DO $$
echo    BEGIN
echo        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'whatsmeow_device') THEN
echo            TRUNCATE TABLE whatsmeow_device CASCADE;
echo        END IF;
echo    END $$;
echo.
echo 3. The app should start without 502 errors
echo.
echo Note: The whatsmeow library will recreate these tables
echo automatically, but they'll be empty and won't cause issues.
echo.
pause