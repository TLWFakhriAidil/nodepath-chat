@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo === Pushing broadcast worker fix ===
echo.

git add src/infrastructure/broadcast/basic_manager.go
git add src/repository/broadcast_repository.go
git commit -m "fix: Broadcast workers now process ALL pending messages

- Fixed processQueueBatch to check ALL pending messages, not just active workers
- Added GetAllPendingMessages method to repository
- Workers are now created on-demand when messages are queued
- Messages queued by campaigns will now be processed automatically"

git push origin main

echo.
echo === Fix pushed! ===
echo.
echo Your messages should now be sent automatically!
echo The system checks for pending messages every 5 seconds.
pause
