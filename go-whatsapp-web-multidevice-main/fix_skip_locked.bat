@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository

REM Create backup
copy broadcast_repository.go broadcast_repository.go.bak

REM Replace FOR UPDATE SKIP LOCKED with just FOR UPDATE for MySQL 5.7
powershell -Command "(Get-Content broadcast_repository.go) -replace 'FOR UPDATE SKIP LOCKED', 'FOR UPDATE' | Set-Content broadcast_repository.go"

echo Fixed FOR UPDATE SKIP LOCKED for MySQL 5.7 compatibility