@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo ========================================
echo Last 20 commits:
echo ========================================
git log --pretty=format:"%%h - %%ad - %%an : %%s" --date=short -20
echo.
echo.
echo ========================================
echo Commits from last 24 hours:
echo ========================================
git log --pretty=format:"%%h - %%ad %%at - %%an : %%s" --date=local --since="24 hours ago"
echo.
pause
