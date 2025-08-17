@echo off
echo Building WhatsApp Multi-Device without CGO...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [1/3] Setting environment variables...
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64

echo [2/3] Building application...
cd src
go build -ldflags="-s -w" -o ../whatsapp.exe main.go

if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Build failed!
    pause
    exit /b 1
)

echo [3/3] Build successful!
echo.
echo Executable created: whatsapp.exe
echo.
echo To test locally: whatsapp.exe rest
echo.
pause
