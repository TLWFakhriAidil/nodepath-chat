@echo off
echo Pushing MySQL fixes to GitHub...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [1/4] Adding changes...
git add -A

echo [2/4] Committing changes...
git commit -m "Fix PostgreSQL syntax for MySQL compatibility - removed ::TIMESTAMP casting, string concatenation with ||, and E'' literals"

echo [3/4] Pulling latest changes...
git pull origin main --no-edit

echo [4/4] Pushing to GitHub...
git push origin main

echo.
echo Done! MySQL fixes have been pushed to GitHub.
echo.
pause
