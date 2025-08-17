@echo off
echo Applying MySQL 5.7 Fix for GetPendingMessagesAndLock...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

REM Backup original file
copy src\repository\broadcast_repository.go src\repository\broadcast_repository.go.backup_before_mysql57_fix

REM Apply the fix
echo.
echo This script will fix the duplicate message issue for MySQL 5.7
echo.
echo The issue: FOR UPDATE SKIP LOCKED is not supported in MySQL 5.7
echo The fix: Use UPDATE-then-SELECT pattern for atomic message claiming
echo.
echo You need to manually replace the GetPendingMessagesAndLock function in:
echo src\repository\broadcast_repository.go
echo.
echo With the version in:
echo src\repository\broadcast_repository_mysql57_fix.go
echo.
echo The key changes:
echo 1. Remove the transaction and FOR UPDATE SKIP LOCKED
echo 2. First UPDATE to claim messages atomically
echo 3. Then SELECT the claimed messages
echo 4. This ensures only one worker can claim each message
echo.
echo After applying the fix:
echo 1. Rebuild: go build -o whatsapp.exe
echo 2. Test the system
echo 3. Check that processing_worker_id is being set correctly
echo.
pause
