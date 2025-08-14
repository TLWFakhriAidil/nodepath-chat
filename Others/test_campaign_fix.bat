@echo off
echo Testing Multiple Campaigns Per Date...
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Building and running the application...
go build -o whatsapp.exe ./src
start whatsapp.exe

echo.
echo The application will now start and automatically:
echo 1. Remove the unique constraint on campaigns table
echo 2. Allow multiple campaigns per date
echo.
echo You can test by:
echo 1. Going to the Campaign tab in the dashboard
echo 2. Click on any date
echo 3. Create multiple campaigns for the same date
echo.
echo The calendar should now show all campaigns with badges!
echo.
pause
