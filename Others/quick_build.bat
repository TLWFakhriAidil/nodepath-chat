@echo off
echo Building without CGO...
cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"
set CGO_ENABLED=0
go build -o ../whatsapp.exe .
echo Build complete!
