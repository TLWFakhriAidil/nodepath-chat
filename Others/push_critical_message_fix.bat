@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "CRITICAL FIX: Messages were not being appended to array in GetPendingMessages" -m "- Fixed missing 'messages = append(messages, msg)' in broadcast_repository.go" -m "- This caused GetPendingMessages to always return empty array" -m "- Messages were being read from DB but never added to return array" -m "- This explains why only greeting was shown - content was empty" -m "- Also added debug logging to track message content flow"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
