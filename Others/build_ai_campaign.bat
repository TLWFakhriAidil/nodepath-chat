@echo off
echo Building WhatsApp Multi-Device System without CGO...
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src
set CGO_ENABLED=0
go build -o ..\whatsapp.exe .
echo Build completed!
cd ..