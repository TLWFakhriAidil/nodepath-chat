@echo off
echo ========================================
echo Git Push Helper for nodepath-chat
echo ========================================
echo.

cd /d "C:\Users\User\Documents\Trae\nodepath-chat-1"

echo Current directory: %CD%
echo.

echo Checking Git status...
git status
echo.

set /p message="Enter commit message (or press Ctrl+C to cancel): "

echo.
echo Adding all changes...
git add .

echo.
echo Committing with message: "%message%"
git commit -m "%message%"

echo.
echo Pushing to GitHub (origin/main)...
git push origin main

echo.
echo ========================================
echo Push completed successfully!
echo ========================================
echo.

pause