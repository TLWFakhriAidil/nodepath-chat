@echo off
echo Fixing sequence contact issues for 3000 devices...

REM Navigate to the main directory
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Step 1: Backing up current code...
copy src\usecase\sequence_trigger_processor.go src\usecase\sequence_trigger_processor.go.backup2

echo.
echo Step 2: Building project...
cd src
go build -o ..\whatsapp_seq_fixed.exe

if %ERRORLEVEL% NEQ 0 (
    echo Build failed! Check errors above.
    pause
    exit /b 1
)

echo.
echo Build successful!
echo.
echo The fixes applied:
echo 1. Database constraints to prevent duplicate active steps
echo 2. Indexes for fast concurrent access by 3000 devices  
echo 3. FOR UPDATE SKIP LOCKED to prevent race conditions
echo 4. Activation by earliest trigger time (not step number)
echo 5. Proper transaction isolation levels
echo.
echo Next steps:
echo 1. Test with your sequences
echo 2. Monitor logs for proper step progression
echo 3. Push to GitHub when confirmed working
echo.
pause