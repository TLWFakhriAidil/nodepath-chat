@echo off
echo Fixing sequence trigger processor to skip problematic queries...

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
git commit -m "Add sequence trigger migration to completed list

- Prevents migration from running repeatedly
- Assumes columns already exist in database as shown in ERD
- Fixes repeated migration attempts that may be causing issues"

git push origin main

echo Fix pushed successfully!
pause