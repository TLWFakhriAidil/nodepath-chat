@echo off
echo ========================================
echo Pushing Bulk Operations Update to GitHub
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add all changes
git add -A

REM Create commit message in a file
echo Add checkbox selection and bulk operations for leads > commit_msg.txt
echo. >> commit_msg.txt
echo - Added checkbox for each lead card with selection state >> commit_msg.txt
echo - Added Select All checkbox to select all visible leads >> commit_msg.txt
echo - Added bulk actions toolbar that appears when leads are selected >> commit_msg.txt
echo - Added bulk delete functionality to delete multiple leads at once >> commit_msg.txt
echo - Added bulk update modal to update niche, target_status, or trigger >> commit_msg.txt
echo - Selected leads are highlighted with blue background >> commit_msg.txt
echo - Shows count of selected leads in bulk actions bar >> commit_msg.txt
echo - All bulk operations use parallel API calls for performance >> commit_msg.txt

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
