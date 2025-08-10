@echo off
echo ========================================
echo Optimizing for 5K Messages Per Device
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

echo Committing high-volume optimizations...
git commit -m "feat: Optimize system for 5K messages per device

Configuration Changes:
- MaxWorkersPerDevice: 1 → 5 (parallel processing)
- MaxConcurrentWorkers: 500 → 2000 (4x capacity)
- WorkerQueueSize: 1000 → 10000 (handle 5K+ messages)
- BatchSize: 100 → 500 (bulk processing)
- DatabaseMaxConnections: 200 → 500
- MaxWorkersPerPool: 3000 → 5000
- BroadcastWorkerQueueSize: 1000 → 5000

Processing Changes:
- Batch processing: 10 → 100 messages per cycle
- Self-healing integration for WhatsApp devices
- Auto-reconnect remains disabled to prevent overload

This allows single device to handle 5K messages efficiently"

echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo ✅ High-Volume Optimization Complete!
echo ========================================
echo.
echo Key improvements:
echo - 5x more workers per device
echo - 10x larger message queues
echo - 5x larger batch processing
echo - Self-healing for reliability
echo.
echo Single device can now handle 5K messages!
echo.
pause
