@echo off
echo Fixing sequence trigger queries to avoid column reference issues...

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
git commit -m "Fix sequence trigger query to avoid CTE column reference issues

- Simplified enrollLeadsFromTriggers query to avoid CTE complexity
- Direct JOIN instead of WITH clauses
- Should fix 'column l.trigger does not exist' error
- Query now directly references leads table columns"

git push origin main

echo Query fixes pushed successfully!
pause