@echo off
echo Checking current status...

git status
echo.

echo Current branch:
git branch --show-current
echo.

echo Attempting to push to origin/main...
git push origin HEAD:main --force -v

echo.
echo If push is hanging, it might be waiting for credentials.
echo You can also try:
echo   1. git remote set-url origin git@github.com:aqilrvsb/Was-MCP.git (for SSH)
echo   2. Or ensure you're logged in to GitHub in your browser
echo.
pause
