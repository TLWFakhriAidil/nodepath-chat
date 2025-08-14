@echo off
echo Fixing schema mismatches based on actual database structure...

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
git commit -m "Fix schema mismatches based on actual database ERD

- Changed current_day to current_step in sequence_contacts queries
- Removed CurrentDay field from SequenceContact model
- Fixed contactJob struct to use currentStep instead of currentDay
- Database already has correct columns, code now matches"

git push origin main

echo Schema fixes pushed successfully!
pause