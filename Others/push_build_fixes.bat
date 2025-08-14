@echo off
echo Building and pushing build fixes...
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src
set CGO_ENABLED=0
go build -o ..\whatsapp.exe .
if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    pause
    exit /b 1
)
cd ..
echo Build successful!
git add -A
git commit -m "Fix build errors in whatsapp_sender.go

- Fixed field names: Url -> URL, FileEncSha256 -> FileEncSHA256, etc
- Added DownloadMedia and GetRandomDelay helper functions
- Removed unused repository import
- Fixed all compilation errors"

git push origin main
echo Push complete!
pause
