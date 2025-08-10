@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix PostgreSQL migration: remove SQLite syntax, fix duplicate methods, update repository signatures"
git push origin main
echo.
echo Push completed!
pause
