@echo off
echo ========================================
echo Pushing Import Lead Update to GitHub
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add all changes
git add -A

REM Create commit message in a file to avoid issues with multi-line
echo Update import lead functionality to only accept 5 columns > commit_msg.txt
echo. >> commit_msg.txt
echo - Updated import modal to show only: name, phone, niche, target_status, trigger >> commit_msg.txt
echo - Made niche and target_status required fields >> commit_msg.txt
echo - Added frontend validation for all required fields >> commit_msg.txt
echo - Updated export to only include these 5 columns >> commit_msg.txt
echo - Removed support for additional_note and device_id columns >> commit_msg.txt
echo - All imported leads now use current device ID >> commit_msg.txt
echo - Added proper validation messages and warnings >> commit_msg.txt

REM Commit using the message file
git commit -F commit_msg.txt

REM Delete the temporary file
del commit_msg.txt

REM Push to main branch
echo.
echo Pushing to GitHub main branch...
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause
