@echo off
echo === Building and Pushing Sequence Fixes ===
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo 1. Building with sequence fixes...
call build_local.bat

if errorlevel 1 (
    echo Build failed!
    pause
    exit /b 1
)

echo.
echo 2. Committing changes...
git add -A
git commit -m "Fix sequence system: Add sequence_stepid tracking and complete monitoring

- Added sequenceStepID field to contactJob struct
- Updated queries to include sequence_stepid
- Added SequenceStepID to BroadcastMessage domain
- Updated QueueMessage to save sequence_stepid
- Enhanced monitorBroadcastResults to sync both success and failure
- Added sequence_failed status for sequences with multiple failures

This allows proper tracking of which specific step's message was sent/failed"

echo.
echo 3. Pushing to GitHub...
git push origin main

echo.
echo === Complete! ===
echo Sequence fixes have been pushed to GitHub.
echo The Go application should now properly track sequence steps.
pause
