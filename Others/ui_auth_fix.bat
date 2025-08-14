@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix auth issues and improve Devices tab UI - modern card design, empty state, better UX"
git push origin main
echo.
echo ============================================
echo UI IMPROVEMENTS & AUTH FIX COMPLETE!
echo ============================================
echo.
echo Changes made:
echo 1. Added /app endpoints to public routes for WhatsApp functionality
echo 2. Added debug logging to auth middleware
echo 3. Completely redesigned Devices tab:
echo    - Modern card layout (2 columns)
echo    - Better visual hierarchy
echo    - Empty state for no devices
echo    - Improved phone number management
echo    - Connected/disconnected visual states
echo    - Better action buttons
echo.
echo The 401 errors should now be fixed!
echo.
pause