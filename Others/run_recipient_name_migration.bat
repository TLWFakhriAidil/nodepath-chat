@echo off
echo Running migration to add recipient_name column...
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

REM You'll need to run this SQL against your database
echo Please run the following SQL in your database:
echo.
type add_recipient_name_column.sql
echo.
echo Migration file created: add_recipient_name_column.sql
pause
