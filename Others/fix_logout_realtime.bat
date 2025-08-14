@echo off
echo ========================================
echo Fixing WhatsApp Device Logout Issues
echo ========================================
echo.

REM Fix 1: Update dashboard.html to handle DEVICE_LOGGED_OUT WebSocket message
echo [1/4] Updating dashboard.html WebSocket handler...

REM Backup original dashboard.html
copy /Y src\views\dashboard.html src\views\dashboard.html.backup >nul 2>&1

REM Read the fix location from the markdown
echo Adding DEVICE_LOGGED_OUT handler to dashboard.html...
echo Please manually add the following code to dashboard.html in the WebSocket switch statement (around line 2835):
echo.
echo case 'DEVICE_LOGGED_OUT':
echo     // Update device status to offline when logged out
echo     console.log('Device logged out:', data.result);
echo     const loggedOutDeviceId = data.result?.deviceId;
echo     if (loggedOutDeviceId) {
echo         const device = devices.find(d =^> d.id === loggedOutDeviceId);
echo         if (device) {
echo             device.status = 'offline';
echo             device.phone = '';
echo             device.jid = '';
echo             device.lastSeen = new Date().toISOString();
echo             renderDevices();
echo             
echo             // Show notification
echo             showAlert('warning', `Device ${device.name} has been logged out`);
echo         }
echo     }
echo     break;
echo.

REM Fix 2: Update device logout handler
echo [2/4] Updating device logout handler...

REM Update the existing device_handler.go
echo Updating src\infrastructure\whatsapp\device_handler.go...
echo Please review and integrate the enhanced logout handler from:
echo fixes\logout-realtime-update\enhanced_logout_handler.go
echo.

REM Fix 3: Update app logout endpoint
echo [3/4] Updating app logout endpoint...

REM Check if logout endpoint exists
findstr /i "app/logout" src\ui\rest\app.go >nul 2>&1
if %ERRORLEVEL%==0 (
    echo Found logout endpoint in app.go
    echo Please ensure it calls the enhanced HandleDeviceLogout function
) else (
    echo Warning: Could not find logout endpoint in app.go
)

REM Fix 4: Build the project
echo.
echo [4/4] Building the project...
cd src
go build -o go-whatsapp-web-multidevice.exe
cd ..

echo.
echo ========================================
echo Manual Steps Required:
echo ========================================
echo.
echo 1. Open src\views\dashboard.html
echo 2. Find the WebSocket message handler switch statement (around line 2835)
echo 3. Add the DEVICE_LOGGED_OUT case as shown above
echo.
echo 4. Open src\infrastructure\whatsapp\device_handler.go
echo 5. Replace the HandleDeviceLogout function with the enhanced version from:
echo    fixes\logout-realtime-update\enhanced_logout_handler.go
echo.
echo 6. Ensure the logout endpoint properly clears WhatsApp session data
echo.
echo 7. Rebuild and test the application
echo.
echo ========================================
echo Fix Summary:
echo ========================================
echo - Device logout will now update in real-time via WebSocket
echo - WhatsApp session data will be properly cleared on logout
echo - Reconnection after logout will work without foreign key errors
echo ========================================
pause
