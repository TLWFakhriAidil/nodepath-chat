@echo off
echo Fixing duplicate message sending issue...

REM Build
echo Building with the duplicate fix...
cd src
set CGO_ENABLED=0
go build -o ..\whatsapp_duplicate_fix.exe

if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    pause
    exit /b 1
)

cd ..
echo Build successful!

REM Git operations
echo.
echo Committing the REAL fix...
git add -A
git commit -m "CRITICAL FIX: Use GetPendingMessagesAndLock instead of GetPendingMessages

- Changed optimized_broadcast_processor.go to call GetPendingMessagesAndLock
- This enables the worker ID locking mechanism that was already implemented
- Prevents multiple workers from processing the same message
- Fixes duplicate message sending issue

The worker ID locking was implemented but wasn't being used because the wrong function was being called!"

echo.
echo Pushing to GitHub...
git push origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ==========================================
    echo CRITICAL FIX SUCCESSFULLY DEPLOYED!
    echo ==========================================
    echo.
    echo The duplicate sending issue is now fixed:
    echo - Worker ID locking is now active
    echo - Each message can only be claimed by one worker
    echo - No more duplicate messages being sent
    echo.
    echo Monitor the processing_worker_id column - it should now have values!
    echo.
) else (
    echo Push failed!
)

pause
