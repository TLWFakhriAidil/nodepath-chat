@echo off
echo Building locally without CGO...
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src
set CGO_ENABLED=0
go build -o whatsapp.exe
if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)

echo Build successful!
cd ..

echo Adding changes...
git add -A

echo Committing...
git commit -m "Fix build error - use correct type domainSequence.UpdateSequenceRequest"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
