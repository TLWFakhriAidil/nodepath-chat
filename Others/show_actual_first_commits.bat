@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo ========================================
echo Checking Git history from the beginning:
echo ========================================

echo.
echo First 30 commits (oldest to newest):
echo ========================================
git log --oneline --reverse | head -30

echo.
echo.
echo Last 30 commits (newest first):
echo ========================================
git log --oneline -30

echo.
echo.
echo Commits by date (January 2025):
echo ========================================
git log --pretty=format:"%%h - %%ad - %%s" --date=short --since="2025-01-01" --until="2025-01-31"

echo.
pause
