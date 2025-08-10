@echo off
echo Creating and force pushing to main branch...

REM First ensure we're on a branch called main
git checkout -b main

REM Add all changes
git add .

REM Commit with a message
git commit -m "Fix sequence model compilation errors - add missing fields to models" --allow-empty

REM Force push to create/update main branch
git push origin main --force

echo.
echo Push completed!
echo.
pause
