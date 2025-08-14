@echo off
echo Fixing WhatsApp message_secrets table...
echo.

REM Get Railway database URL
echo Getting database connection from Railway...
railway run psql %DATABASE_URL% -f fix_message_secrets_column.sql

echo.
echo Fix applied! The "key" column should now exist in whatsmeow_message_secrets table.
echo.
pause