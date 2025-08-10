@echo off
echo Creating clean team dashboard from master dashboard...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

REM Copy dashboard.html as base
copy src\views\dashboard.html src\views\team_dashboard_clean.html

echo Team dashboard template created. Now need to modify it...
pause