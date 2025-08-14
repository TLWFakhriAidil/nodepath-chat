@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src
echo Building without CGO...
set CGO_ENABLED=0
go build -o ../whatsapp_test.exe .
if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)
echo Build successful!
pause