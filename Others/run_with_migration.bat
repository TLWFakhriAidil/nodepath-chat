@echo off
echo ========================================
echo Running WhatsApp with Auto-Migration
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo The application will now:
echo 1. Connect to PostgreSQL database
echo 2. Run auto-migration from connection.go
echo 3. Fix sequence_steps table structure
echo    - Remove: send_time, created_at, updated_at, day, schedule_time
echo    - Add: trigger, next_trigger, trigger_delay_hours, etc.
echo 4. Start REST API on port 3000
echo.
echo Starting server with debug mode...
echo.

src\whatsapp.exe rest --debug=true --port=3000

pause
