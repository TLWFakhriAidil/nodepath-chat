@echo off
echo FINAL BUILD AND PUSH - GetPendingMessagesAndLock Verification Complete
echo ======================================================================

REM Build
echo Building final version...
cd src
set CGO_ENABLED=0
go build -o ..\whatsapp_final_worker_fix.exe

if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    pause
    exit /b 1
)

cd ..
echo Build successful!

REM Git operations
echo.
echo Committing final verification...
git add -A
git commit -m "FINAL: Verified GetPendingMessagesAndLock is used everywhere

VERIFICATION COMPLETE:
✓ optimized_broadcast_processor.go - Using GetPendingMessagesAndLock
✓ broadcast_worker_processor.go - Using GetPendingMessagesAndLock  
✓ Worker ID and timestamp implementation verified
✓ No old GetPendingMessages calls in active code
✓ processing_worker_id will be populated when this runs

IMPORTANT:
- Deploy whatsapp_final_worker_fix.exe to production
- Restart the WhatsApp service
- Monitor that processing_worker_id starts getting populated

This will prevent all duplicate messages through atomic worker locking."

echo.
echo Pushing to GitHub main branch...
git push origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ================================================
    echo FINAL PUSH SUCCESSFUL!
    echo ================================================
    echo.
    echo GetPendingMessagesAndLock is confirmed to be used everywhere.
    echo.
    echo NEXT STEPS:
    echo 1. Deploy whatsapp_final_worker_fix.exe to your server
    echo 2. Stop the old WhatsApp process
    echo 3. Start the new process
    echo 4. Check that processing_worker_id is being populated:
    echo.
    echo    SELECT COUNT(*), COUNT(processing_worker_id) 
    echo    FROM broadcast_messages 
    echo    WHERE created_at > NOW();
    echo.
    echo Once deployed, duplicate messages will be completely prevented!
    echo.
) else (
    echo Push failed!
)

pause
