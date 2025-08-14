@echo off
echo Fixing Go code to match actual database schema...

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
git commit -m "Fix Go code to match actual database schema

- Removed s.priority from ORDER BY clause (column doesn't exist)
- Code now matches actual database structure shown in ERD
- No database changes needed - just fixed Go queries"

git push origin main

echo Go code fixes pushed successfully!
pause