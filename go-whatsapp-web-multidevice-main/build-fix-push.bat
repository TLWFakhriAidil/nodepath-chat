@echo off
echo ===================================
echo WhatsApp Multi-Device Build Script
echo ===================================

echo.
echo Step 1: Building the application...
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src

REM Try to build with Go
go build -o whatsapp.exe main.go

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Build failed. You need to install GCC first.
    echo.
    echo Please install MinGW-w64:
    echo 1. Download from: https://github.com/niXman/mingw-builds-binaries/releases
    echo 2. Extract to C:\mingw64
    echo 3. Add C:\mingw64\bin to PATH
    echo 4. Restart this script
    echo.
    pause
    exit /b 1
)

echo Build successful!

echo.
echo Step 2: Committing changes...
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

git add .
git commit -m "Fix: Add support for data URL images in broadcast messages

- Updated downloadMedia function to handle base64 encoded data URLs
- Added base64 and strings imports
- Now supports both HTTP URLs and data URLs for images"

echo.
echo Step 3: Pushing to remote...
git push origin main

if %ERRORLEVEL% EQ 0 (
    echo.
    echo ===================================
    echo SUCCESS! Changes pushed to GitHub
    echo ===================================
) else (
    echo.
    echo ERROR: Failed to push. Please check your git credentials.
)

pause
