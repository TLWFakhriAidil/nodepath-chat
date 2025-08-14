@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo ================================================
echo Checking database changes since commit b07cfad
echo ================================================
echo.
echo Files changed related to database:
git diff b07cfad HEAD -- "*.sql" "*migration*" "*database*" "*connection.go"
echo.
pause