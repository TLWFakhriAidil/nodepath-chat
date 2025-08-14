@echo off
echo Pushing message sending logic fixes...
echo.

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding changes...
git add -A

echo.
echo Committing...
git commit -m "Fix message sending logic with proper delays and two-part messages

CRITICAL FIXES:
1. Two-part messages (image + text):
   - Send image first WITHOUT caption
   - Wait 3 seconds
   - Then send text message
   
2. Random delay between leads:
   - Use min/max delay from campaign/sequence
   - Random value between min and max
   - Applied AFTER each lead (not between image/text)
   
3. Single worker for both campaigns and sequences:
   - Same worker handles BOTH types
   - Shared queue for optimal resource usage
   - Prevents duplicate workers per device
   
4. Per-message delay settings:
   - Each message carries its own min/max delay
   - Campaigns use campaign delays
   - Sequences use step-specific delays
   
This ensures proper WhatsApp-like behavior and prevents detection."

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo Done! Message sending logic has been fixed.
pause
