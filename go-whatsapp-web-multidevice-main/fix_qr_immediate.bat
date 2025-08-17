@echo off
echo ========================================
echo WhatsApp Multi-Device QR Code Quick Fix
echo ========================================
echo.

echo Step 1: Killing any running WhatsApp processes...
taskkill /F /IM whatsapp.exe 2>nul
timeout /t 2

echo.
echo Step 2: Clearing QR code cache...
if exist "src\views\qrcode" (
    del /Q "src\views\qrcode\*.png" 2>nul
    echo QR code cache cleared.
) else (
    echo No QR cache directory found.
)

echo.
echo Step 3: Creating QR directory if missing...
if not exist "src\views\qrcode" (
    mkdir "src\views\qrcode"
    echo QR directory created.
)

echo.
echo Step 4: Setting environment variables...
set APP_DEBUG=true
set WHATSAPP_CHAT_STORAGE=true

echo.
echo Step 5: Starting application with verbose logging...
echo.
echo ========================================
echo IMPORTANT: Try these methods to connect:
echo.
echo METHOD 1 - QR Code:
echo 1. Click "Add Device" in dashboard
echo 2. Wait 5-10 seconds for QR to generate
echo 3. Scan with WhatsApp mobile app
echo.
echo METHOD 2 - Phone Code (More Reliable):
echo 1. Click "Phone Code" button
echo 2. Enter phone with country code: +60123456789
echo 3. Get 8-character code
echo 4. In WhatsApp: Settings > Linked Devices > Link with phone number
echo 5. Enter the 8-character code
echo ========================================
echo.
echo Starting application now...
cd src
go run main.go

pause
