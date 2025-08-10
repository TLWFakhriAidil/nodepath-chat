@echo off
echo Removing debug logs for cleaner production output...

REM Build without CGO
echo Building application...
cd src
set CGO_ENABLED=0
go build -o ../whatsapp.exe

if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)

echo Build successful!
cd ..

REM Commit and push
git add -A
git commit -m "Remove debug logs for cleaner production output

- Commented out GetUserByEmail debug logs
- Commented out password validation debug logs  
- Commented out API Auth Debug logs
- Commented out session validation logs
- Commented out login attempt/success logs
- Cleaner logs without sensitive information exposure"

git push origin main

echo Debug logs removal pushed successfully!
pause