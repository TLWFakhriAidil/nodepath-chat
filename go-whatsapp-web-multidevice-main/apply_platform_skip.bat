@echo off
echo =============================================
echo Adding Platform Skip Feature to Device Checks
echo =============================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

echo Step 1: Creating backup of original files...
if not exist backups\platform_skip mkdir backups\platform_skip

copy src\infrastructure\whatsapp\device_status_normalizer.go backups\platform_skip\ >nul 2>&1
copy src\infrastructure\whatsapp\auto_connection_monitor_15min.go backups\platform_skip\ >nul 2>&1
copy src\usecase\optimized_campaign_trigger.go backups\platform_skip\ >nul 2>&1
copy src\usecase\sequence_trigger_processor.go backups\platform_skip\ >nul 2>&1

echo Step 2: Applying platform skip updates...

REM Replace the original files with platform skip versions
copy src\infrastructure\whatsapp\device_status_normalizer_skip_platform.go src\infrastructure\whatsapp\device_status_normalizer.go >nul
copy src\infrastructure\whatsapp\auto_connection_monitor_15min_skip_platform.go src\infrastructure\whatsapp\auto_connection_monitor_15min.go >nul

echo Step 3: Creating SQL update script...
echo -- Add platform column and update queries > update_platform_skip.sql
echo. >> update_platform_skip.sql
type add_platform_column.sql >> update_platform_skip.sql

echo.
echo Platform skip feature has been prepared!
echo.
echo Next steps:
echo 1. Run the SQL script: update_platform_skip.sql
echo 2. Update the sequence processor query manually
echo 3. Build and test the application
echo.
pause
