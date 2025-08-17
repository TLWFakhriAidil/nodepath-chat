@echo off
echo Fixing campaign trigger issue...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src

echo.
echo Running SQL to check and fix campaign status...

set PGPASSWORD=your_password_here

psql -U postgres -d your_database -c "UPDATE campaigns SET status = 'pending' WHERE title = 'test' AND status = 'scheduled';"
psql -U postgres -d your_database -c "SELECT id, title, status, campaign_date, scheduled_time FROM campaigns WHERE title = 'test';"

echo.
echo Campaign status updated. The trigger service should pick it up within 1 minute.
echo Check Worker Status page to monitor progress.
pause