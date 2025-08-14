@echo off
echo ========================================
echo Railway Deployment Tester
echo ========================================
echo.

set /p RAILWAY_URL=Enter your Railway app URL (e.g., https://your-app.up.railway.app): 
set /p AUTH_USER=Enter username (default: admin): 
set /p AUTH_PASS=Enter password (default: changeme123): 

if "%AUTH_USER%"=="" set AUTH_USER=admin
if "%AUTH_PASS%"=="" set AUTH_PASS=changeme123

echo.
echo Testing %RAILWAY_URL% with user: %AUTH_USER%
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\testing

REM Create temporary Python script with custom URL
echo import sys > temp_test.py
echo sys.path.insert(0, '.') >> temp_test.py
echo from test_railway_live import WhatsAppSystemTester >> temp_test.py
echo import test_railway_live >> temp_test.py
echo. >> temp_test.py
echo # Override configuration >> temp_test.py
echo test_railway_live.RAILWAY_URL = "%RAILWAY_URL%" >> temp_test.py
echo test_railway_live.API_BASE = f"{test_railway_live.RAILWAY_URL}/api/v1" >> temp_test.py
echo test_railway_live.AUTH = ("%AUTH_USER%", "%AUTH_PASS%") >> temp_test.py
echo. >> temp_test.py
echo # Run tests >> temp_test.py
echo tester = WhatsAppSystemTester() >> temp_test.py
echo tester.run_all_tests() >> temp_test.py

python temp_test.py

del temp_test.py

pause
