@echo off
echo Building and pushing sequence modal date filter fix...

REM Build without CGO
echo Building application...
set CGO_ENABLED=0
go build -o whatsapp_sequence_modal_fix.exe

if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    pause
    exit /b 1
)

echo Build successful!

REM Git operations
echo.
echo Adding changes to git...
git add -A

echo.
echo Committing changes...
git commit -m "Fix: Add date filter to sequence step leads modal

- Added start_date and end_date query parameters to GetSequenceStepLeads API
- Updated showSequenceStepLeadDetails JS function to pass current date filters
- Modal now respects the selected date range instead of showing all messages
- Fixes issue where clicking success count showed messages from all dates"

echo.
echo Pushing to GitHub...
git push origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo SUCCESS! Changes pushed to GitHub
    echo ========================================
    echo.
    echo The fix has been applied for:
    echo - Sequence step leads modal now respects date filters
    echo - When filtering for a specific date, only shows messages from that date
    echo.
) else (
    echo.
    echo Push failed! Please check your git configuration.
)

pause
