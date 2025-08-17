@echo off
echo Building without CGO...
cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"

REM Set environment variable properly for Windows
set CGO_ENABLED=0

REM Build
go build -o ../whatsapp.exe .

echo Build complete!
