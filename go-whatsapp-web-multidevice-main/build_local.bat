@echo off
REM Build for local development WITHOUT CGO

echo ========================================
echo Building WhatsApp Multi-Device (Local)
echo WITHOUT CGO_ENABLED
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"

REM Clean previous build
if exist ..\whatsapp.exe (
    echo Removing old build...
    del ..\whatsapp.exe
)

REM Build WITHOUT CGO for local
echo Building application (CGO_ENABLED=0)...
set CGO_ENABLED=0
go build -o ../whatsapp.exe .

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Build failed!
    pause
    exit /b 1
)

echo.
echo Build successful!
echo.

REM Navigate back and run
cd ..
echo Starting WhatsApp Multi-Device System...
echo.
whatsapp.exe

pause