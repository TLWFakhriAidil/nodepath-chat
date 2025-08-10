@echo off
echo Force pushing to main branch...

git add -A
git commit -m "Fix sequence model compilation errors - add missing fields" --allow-empty
git push origin master:main --force

echo.
echo Force push to main branch completed!
echo.
pause
