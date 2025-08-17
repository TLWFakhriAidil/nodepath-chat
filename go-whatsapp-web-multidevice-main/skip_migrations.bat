@echo off
echo Disabling auto-migrations since database is already configured...

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
git commit -m "Skip auto-migrations - database schema already configured

- Commented out migration code in connection.go
- Database already has all required tables and columns
- Prevents unnecessary migration attempts on startup
- Improves startup time and reduces error logs"

git push origin main

echo Migration skip pushed successfully!
pause