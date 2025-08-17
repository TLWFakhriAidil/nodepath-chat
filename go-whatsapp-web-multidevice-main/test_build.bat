@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src
echo Building...
go build -o ../whatsapp_test.exe .
if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)
echo Build successful!
pause