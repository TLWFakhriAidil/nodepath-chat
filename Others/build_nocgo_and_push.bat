@echo off
echo === Building WhatsApp Multi-Device without CGO ===
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Setting CGO_ENABLED=0...
set CGO_ENABLED=0

echo Building...
go build -o whatsapp_nocgo.exe ./src

if %errorlevel% neq 0 (
    echo.
    echo Build failed! Checking for errors...
    pause
    exit /b 1
)

echo.
echo Build successful!
echo Executable: whatsapp_nocgo.exe
echo.
echo === Committing and pushing to GitHub ===
git add src/usecase/direct_broadcast_processor.go
git commit -m "Fix sequence SQL syntax - properly escape trigger keyword in queries"
git push origin main

echo.
echo === Complete ===
pause
