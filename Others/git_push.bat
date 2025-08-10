@echo off
echo ====================================
echo Initializing Git and Pushing to GitHub
echo ====================================

REM Initialize git if not already initialized
if not exist .git (
    echo Initializing Git repository...
    git init
    echo Git repository initialized!
) else (
    echo Git repository already exists.
)

REM Add all files
echo Adding all files to Git...
git add .

REM Commit with message
echo Committing changes...
git commit -m "Add Ultra Stable Connection Mode - Devices never disconnect, maximum speed messaging"

REM Check if remote exists
git remote -v | find "origin" >nul
if errorlevel 1 (
    echo No remote origin found.
    echo Please add your GitHub repository:
    echo Example: git remote add origin https://github.com/yourusername/go-whatsapp-web-multidevice.git
    echo.
    echo After adding remote, run: git push -u origin main
) else (
    echo Remote origin found. Pushing to main branch...
    git push origin main
    
    if errorlevel 1 (
        echo.
        echo Push failed. Trying to push to main branch with upstream...
        git push -u origin main
    )
)

echo ====================================
echo Done!
echo ====================================
