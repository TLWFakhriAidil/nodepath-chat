@echo off
echo Committing device tracking improvements...
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

git add src/usecase/sequence_trigger_processor.go
git commit -m "Keep processing_device_id for tracking and maintain strict device ownership

- Keep processing_device_id after completion for tracking which device processed
- Remove stuck processing cleanup per request
- Don't release contact on failure - maintain strict device ownership
- Only the assigned device can process its leads
- Better tracking of which device handled each sequence step"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo Done! Changes pushed to GitHub.
pause