@echo off
echo Fixing sequence model compilation errors...

git add -A
git commit -m "Fix sequence model compilation errors - add missing fields"
git push origin main --force

echo.
echo Fix pushed successfully!
echo.
pause
