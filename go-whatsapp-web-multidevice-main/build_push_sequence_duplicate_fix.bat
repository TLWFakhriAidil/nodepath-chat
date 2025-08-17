@echo off
echo Fixing sequence duplicate messages...

REM Build with the fix
echo Building application with duplicate fix...
cd src
set CGO_ENABLED=0
go build -o ..\whatsapp_sequence_duplicate_fix.exe

if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    pause
    exit /b 1
)

echo Build successful!
cd ..

REM Git operations
echo.
echo Committing changes...
git add -A
git commit -m "Fix: Prevent sequence duplicate messages

- Added transaction support to QueueMessage for atomic duplicate checking
- Added 'processing' status to duplicate check conditions
- Fixed race condition where multiple processes could create duplicates
- Ensures only one message per sequence step/phone/device combination"

echo.
echo Pushing to GitHub...
git push origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo SUCCESS! Duplicate fix pushed to GitHub
    echo ========================================
    echo.
    echo The fix addresses:
    echo - Race condition in message creation
    echo - Atomic duplicate checking with transactions
    echo - Prevents multiple processes from creating same message
    echo.
    echo IMPORTANT: Also run this SQL to add unique constraint:
    echo.
    echo ALTER TABLE broadcast_messages 
    echo ADD UNIQUE INDEX IF NOT EXISTS unique_sequence_message (
    echo     sequence_stepid, 
    echo     recipient_phone, 
    echo     device_id
    echo );
    echo.
) else (
    echo.
    echo Push failed! Please check your git configuration.
)

pause
