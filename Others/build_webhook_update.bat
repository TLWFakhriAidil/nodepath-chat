@echo off
echo Building WhatsApp Multi-Device with updated webhook logic...
echo.

REM Navigate to src directory
cd src

REM First backup the original webhook file
echo Backing up original webhook file...
copy ui\rest\webhook_lead.go ui\rest\webhook_lead_backup.go

REM Replace with updated version
echo Applying updated webhook logic...
copy /Y ui\rest\webhook_lead_updated.go ui\rest\webhook_lead.go

REM Add the new repository function
echo Adding GetDeviceByUserAndName function to repository...
type repository\user_repository_addition.go >> repository\user_repository.go

REM Build without CGO for Windows
echo Building application without CGO...
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64
go build -o ../whatsapp_updated.exe main.go

REM Check if build was successful
if %ERRORLEVEL% EQU 0 (
    echo.
    echo Build successful! Output: whatsapp_updated.exe
    echo.
    echo Changes applied:
    echo - Webhook now checks for existing devices by device_name first
    echo - If device exists, only JID is updated
    echo - Prevents duplicate devices with same name
    echo.
) else (
    echo.
    echo Build failed! Please check the error messages above.
    echo Restoring original webhook file...
    copy /Y ui\rest\webhook_lead_backup.go ui\rest\webhook_lead.go
)

cd ..

echo.
echo Next steps:
echo 1. Test the webhook with same device_name multiple times
echo 2. Verify only JID gets updated, not creating new devices
echo 3. Push to GitHub when ready
pause