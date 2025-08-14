@echo off
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src
echo Building with CGO disabled...
set CGO_ENABLED=0
go build -v -x -o ..\whatsapp.exe . 2>&1
if %ERRORLEVEL% neq 0 (
    echo Build failed with error code %ERRORLEVEL%
) else (
    echo Build successful!
)
pause