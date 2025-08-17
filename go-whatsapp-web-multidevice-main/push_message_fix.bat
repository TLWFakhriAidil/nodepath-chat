@echo off
echo Building and pushing message processing fix...
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src
set CGO_ENABLED=0
go build -o ..\whatsapp.exe .
if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    pause
    exit /b 1
)
cd ..
echo Build successful!
git add -A
git commit -m "Fix: Message processing - bridge Redis queue to worker queue

- Fixed processMessage to queue to worker's internal queue
- Worker now properly processes messages from its queue
- Added whatsapp_sender.go with detailed message sending logic
- Messages should now actually be sent via WhatsApp
- Fixed the disconnect between Redis queue and worker processing"

git push origin main
echo Push complete!
echo.
echo FIX SUMMARY:
echo 1. Messages from Redis now properly queued to worker
echo 2. Worker processes messages from its internal queue
echo 3. WhatsApp messages should now be sent
echo 4. Status updates should work correctly
echo.
pause
