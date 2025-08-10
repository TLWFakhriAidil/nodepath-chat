@echo off
echo ========================================
echo Pushing SQL NULL Fix for Broadcast Processor
echo ========================================
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "fix: Handle NULL image_url in broadcast processor

- Fixed SQL scanning error for nullable image_url field
- Changed to use sql.NullString for proper NULL handling
- Prevents 'converting NULL to string is unsupported' error

This fixes the broadcast processor that was failing to process
messages without images."

REM Push to main branch
git push origin main

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo Push failed! You may need to pull first.
    pause
    exit /b 1
)

echo.
echo ========================================
echo Successfully pushed SQL NULL fix!
echo ========================================
echo.
echo Your broadcast messages should now process correctly!
echo.
pause
