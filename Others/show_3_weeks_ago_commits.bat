@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo ========================================
echo Git commits from 3 weeks ago
echo ========================================
echo.

REM Calculate 3 weeks ago (21 days)
echo Showing commits from 21-22 days ago:
echo ========================================
git log --pretty=format:"%%h - %%ad - %%an : %%s" --date=short --since="22 days ago" --until="21 days ago"

echo.
echo.
echo Showing commits from the entire week (21-28 days ago):
echo ========================================
git log --pretty=format:"%%h - %%ad - %%an : %%s" --date=short --since="28 days ago" --until="21 days ago"

echo.
echo.
echo Showing commits from around 3 weeks ago (18-24 days):
echo ========================================
git log --pretty=format:"%%h - %%ad - %%an : %%s" --date=short --since="24 days ago" --until="18 days ago"

echo.
echo.
echo All commits with relative dates:
echo ========================================
git log --pretty=format:"%%h - %%ar - %%an : %%s" --since="30 days ago" --until="15 days ago"

echo.
pause
