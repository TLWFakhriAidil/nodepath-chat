@echo off
echo Committing sequence enrollment fix...
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

git add src/usecase/sequence_trigger_processor.go
git commit -m "Fix sequence enrollment to create all steps with chain activation

- Changed enrollment to create ALL steps at once (not just first step)
- First step is ACTIVE with 5 minute delay, others are PENDING
- Each step's next_trigger_time = previous step time + trigger_delay_hours
- After sending message, complete current step and activate next pending step
- Chain reaction: complete -> activate next -> wait for time -> process
- No separate activation query needed, handled in updateContactProgress
- Added transaction for atomic complete + activate operations
- Better logging with emojis for clear status tracking"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo Done! Changes pushed to GitHub.
pause