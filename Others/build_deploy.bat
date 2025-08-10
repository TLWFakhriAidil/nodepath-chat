@echo off
REM Build for deployment WITH CGO

echo ========================================
echo Building WhatsApp Multi-Device (Deploy)
echo WITH CGO_ENABLED
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"

REM Clean previous build
if exist ..\whatsapp.exe (
    echo Removing old build...
    del ..\whatsapp.exe
)

REM Build WITH CGO for deployment
echo Building application (CGO_ENABLED=1)...
set CGO_ENABLED=1
go build -o ../whatsapp.exe .

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Build failed!
    pause
    exit /b 1
)

echo.
echo Build successful!
echo Ready for deployment!
echo.

pause