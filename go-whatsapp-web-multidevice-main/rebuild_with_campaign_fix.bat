@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src
echo Rebuilding WhatsApp system with campaign fixes...

echo.
echo Building application without CGO...
set CGO_ENABLED=0
go build -o ..\whatsapp.exe main.go

if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    pause
    exit /b 1
)

echo.
echo Build successful!
echo.
echo IMPORTANT: The database schema has been updated.
echo Please restart the application for changes to take effect.
echo.
echo If campaigns still don't show, run this SQL on your database:
echo   ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS device_id UUID;
echo.
pause
