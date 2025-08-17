@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo ========================================
echo Git commits from today:
echo ========================================
git log --pretty=format:"%%h - %%an, %%ar : %%s" --since="today 00:00:00"
echo.
echo ========================================
pause
