@echo off
echo Building and pushing Redis cleanup tools...
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
git commit -m "Add Redis cleanup tools and debug logging for campaigns

- Added Redis cleanup API endpoints
- Created cleanup page at /redis-cleanup
- Enhanced campaign logging to show lead details
- Added tools to remove old device data from Redis
- Better debugging for why campaigns show 0 messages"

git push origin main
echo Push complete!
pause
