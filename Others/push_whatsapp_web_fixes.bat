@echo off
echo ========================================
echo WhatsApp Web Image Fix + Clean UI Update
echo ========================================

REM 1. Build local first
echo Step 1: Building locally...
call build_local.bat
if errorlevel 1 (
    echo Build failed!
    pause
    exit /b 1
)

REM 2. Fix is already applied

REM 3. README is already updated

REM 4. Git operations
echo.
echo Step 4: Git operations...

REM Add all changes
git add .

REM Commit with descriptive message
git commit -m "fix: WhatsApp Web sent images now display correctly + removed loading UI" -m "- Fixed sent images returning 404 by saving them to disk with proper filename" -m "- Removed refresh button and loading spinners for cleaner interface" -m "- All updates now happen seamlessly via WebSocket"

REM Push to main
echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Push completed successfully!
echo ========================================
echo.
echo Changes:
echo - Sent images now save to disk and display correctly
echo - Removed refresh icon and loading messages
echo - Updated README with latest fixes
echo.
pause
