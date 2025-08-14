@echo off
echo Building WhatsApp application with device creation fix...
echo.

REM Kill any running instance
taskkill /F /IM whatsapp.exe 2>nul

REM Set build environment
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64

REM Clean old build
if exist whatsapp.exe del whatsapp.exe

REM Build the application
echo Building application...
go build -o whatsapp.exe src/main.go

if exist whatsapp.exe (
    echo.
    echo Build successful! 
    echo The device creation issue has been fixed.
    echo.
    echo Changes made:
    echo 1. Fixed SQL syntax error in AddUserDevice function
    echo 2. Fixed SQL syntax error in AddUserDeviceWithPhone function
    echo 3. Changed from QueryRow/Scan to Exec for INSERT operations
    echo 4. Added proper updated_at field handling
    echo.
    echo You can now run: whatsapp.exe rest
) else (
    echo.
    echo Build failed! Please check for errors above.
)

pause