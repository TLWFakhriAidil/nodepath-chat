@echo off
echo Running Campaign Database Fix on Railway PostgreSQL...
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Building migration tool...
go build -o migration.exe run_migration.go

echo.
echo Running migration...
echo Please make sure your DB_URI environment variable is set with your Railway PostgreSQL connection string
echo.

migration.exe

echo.
echo Migration complete!
echo.

del migration.exe

pause
