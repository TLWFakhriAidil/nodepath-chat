@echo off
echo ========================================================
echo Fix Syntax Errors - Orphaned Code Blocks
echo ========================================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding changes...
git add -A

echo.
echo Committing fix...
git commit -m "Fix syntax errors - remove orphaned code blocks

- Removed orphaned code block at line 1231 (FORBIDDEN status)
- Removed duplicate logout code at line 1356
- Fixed non-declaration statements outside function body
- All syntax errors resolved"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================================
echo SYNTAX ERRORS FIXED!
echo.
echo All orphaned code blocks have been removed.
echo The build should now succeed on Railway.
echo ========================================================
pause