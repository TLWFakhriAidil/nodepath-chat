@echo off
echo ========================================
echo FIXING COMPILATION ERROR
echo ========================================

echo.
echo Committing compilation fix...
git add -A
git commit -m "fix: Add missing error variable declaration in campaign_trigger.go

- Fixed undefined 'err' variable compilation error
- Added proper error variable declaration"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================
echo FIX DEPLOYED!
echo ========================================
echo.
echo Compilation error fixed!
echo.
pause
