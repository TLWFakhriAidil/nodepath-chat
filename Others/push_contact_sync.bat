@echo off
echo ========================================
echo Pushing WhatsApp Contact Sync Update
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add all changes
git add -A

REM Create commit message
echo Add WhatsApp contact sync with 6-month history > commit_msg.txt
echo. >> commit_msg.txt
echo - Changed chat history from 30 days to 6 months >> commit_msg.txt
echo - Added auto-save WhatsApp contacts to leads with duplicate prevention >> commit_msg.txt
echo - Added sync contacts button in device actions page >> commit_msg.txt
echo - Preserve existing data when re-scanning or changing devices >> commit_msg.txt
echo - Added device merge functionality for banned devices >> commit_msg.txt
echo - All operations use INSERT ON CONFLICT DO NOTHING >> commit_msg.txt
echo - No data is ever deleted, only added >> commit_msg.txt

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
