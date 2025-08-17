@echo off
echo ========================================
echo Implementing Self-Healing for ALL Broadcast Workers
echo ========================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src
set CGO_ENABLED=0

echo Building without CGO...
go build -o whatsapp.exe
if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)

echo Build successful!
echo.

cd ..

echo Copying executable to root...
copy src\whatsapp.exe whatsapp.exe /Y

echo Adding all changes...
git add -A

echo Committing self-healing integration...
git commit -m "fix: Integrate self-healing into broadcast worker processor

- Updated broadcast_worker_processor.go to use WhatsAppMessageSender
- Now ALL WhatsApp devices (platform=null) use self-healing connection refresh
- Platform devices continue using external APIs without refresh
- This fixes 'timeout queueing message to worker' errors
- Each message now gets a fresh, healthy connection before sending"

echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo ✅ Self-Healing Integration Complete!
echo ========================================
echo.
echo Changes made:
echo - broadcast_worker_processor now uses self-healing
echo - WhatsApp devices will refresh connections per message
echo - Platform devices unaffected (use external APIs)
echo.
echo Deploy this to Railway to fix timeout errors!
echo.
pause
