@echo off
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src

echo Initializing git if needed...
git init

echo Adding remote...
git remote add origin https://github.com/RAIZERO-Team/go-whatsapp-web-multidevice.git 2>nul

echo Pulling latest...
git pull origin main --allow-unrelated-histories 2>nul

echo Adding changes...
git add ui/rest/app.go

echo Committing...
git commit -m "Fix sequence device report calculation to match summary page logic"

echo Pushing to GitHub...
git push -u origin main

echo.
echo Done! Check GitHub Actions for deployment status.
pause