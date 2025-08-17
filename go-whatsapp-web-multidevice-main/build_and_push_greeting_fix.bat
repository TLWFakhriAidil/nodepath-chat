@echo off
echo === Building and Pushing Greeting Fix ===
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Building with greeting fix...
cd src
set CGO_ENABLED=0
go build -o ..\whatsapp_greeting_fix.exe .

if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)

cd ..
echo Build successful!
echo.

echo Committing and pushing to GitHub...
git add src/pkg/antipattern/greeting_processor.go
git commit -m "Fix greeting line break - use single line break for platform compatibility"
git push origin main

echo.
echo === Complete ===
echo The fix changes double line break to single line break between greeting and message
echo This should work better with platform APIs like Wablas/Whacenter
pause
