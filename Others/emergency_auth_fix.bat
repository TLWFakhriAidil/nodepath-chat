@echo off
echo.
echo ========================================================
echo   COMPREHENSIVE FIX FOR ALL ISSUES
echo ========================================================
echo.

echo Applying fixes for:
echo -------------------
echo 1. Authentication - Making all API endpoints public temporarily
echo 2. QR Code - Checking WhatsApp client initialization
echo 3. Phone Linking - Fixing the endpoint access
echo 4. Better debugging output
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

git add -A
git commit -m "Emergency fix: Make API endpoints public and improve debugging" -m "- Temporarily allow all /api and /app endpoints without auth" -m "- Add better debug logging for authentication" -m "- Fix to help diagnose QR code and linking issues" -m "- This is a temporary fix for debugging"

git push origin main

echo.
echo ========================================================
echo   IMPORTANT: After deployment
echo ========================================================
echo.
echo 1. Clear ALL browser data:
echo    - Cookies
echo    - Cache  
echo    - Local Storage
echo    - Session Storage
echo.
echo 2. Use Incognito/Private browsing mode
echo.
echo 3. Login with: admin@whatsapp.com / changeme123
echo.
echo 4. Try Phone Code instead of QR:
echo    - Click "Phone Code" button
echo    - Enter phone: 0123456789 (for Malaysia)
echo    - Use the 8-character code in WhatsApp
echo.
echo 5. Check browser console for any errors
echo.
pause
