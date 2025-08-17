@echo off
echo ================================================
echo Disabling ALL Auto-Reconnection Functions
echo ================================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src

echo 1. Disabling Device Health Monitor...
powershell -Command "(Get-Content cmd\rest.go) -replace 'healthMonitor.Start\(\)', '// healthMonitor.Start() // DISABLED - No auto reconnect' | Set-Content cmd\rest.go"

echo 2. Disabling Error Monitor...
powershell -Command "(Get-Content cmd\root.go) -replace 'whatsapp.MonitorDeviceErrors\(\)', '// whatsapp.MonitorDeviceErrors() // DISABLED - No auto reconnect' | Set-Content cmd\root.go"

echo 3. Disabling Auto Connection Monitor...
REM Find where it's initialized and disable it
powershell -Command "(Get-Content infrastructure\whatsapp\auto_connection_monitor_15min.go) -replace 'func \(acm \*AutoConnectionMonitor\) Start\(\) error \{', 'func (acm *AutoConnectionMonitor) Start() error { return nil // DISABLED - No auto reconnect' | Set-Content infrastructure\whatsapp\auto_connection_monitor_15min.go"

echo 4. Disabling Multi-Device Auto Reconnect...
powershell -Command "(Get-Content infrastructure\whatsapp\multidevice_auto_reconnect.go) -replace 'func StartMultiDeviceAutoReconnect\(\) \{', 'func StartMultiDeviceAutoReconnect() { return // DISABLED - No auto reconnect' | Set-Content infrastructure\whatsapp\multidevice_auto_reconnect.go"

echo 5. Disabling Device Health Check Loop...
powershell -Command "(Get-Content infrastructure\whatsapp\device_health_monitor.go) -replace 'func \(dhm \*DeviceHealthMonitor\) monitor\(\) \{', 'func (dhm *DeviceHealthMonitor) monitor() { return // DISABLED - No auto reconnect' | Set-Content infrastructure\whatsapp\device_health_monitor.go"

echo 6. Disabling Reconnect Methods in Device Health Monitor...
powershell -Command "(Get-Content infrastructure\whatsapp\device_health_monitor.go) -replace 'func \(dhm \*DeviceHealthMonitor\) checkDeviceHealth', 'func (dhm *DeviceHealthMonitor) checkDeviceHealth_DISABLED' | Set-Content infrastructure\whatsapp\device_health_monitor.go"
powershell -Command "(Get-Content infrastructure\whatsapp\device_health_monitor.go) -replace 'func \(dhm \*DeviceHealthMonitor\) reconnectDevice', 'func (dhm *DeviceHealthMonitor) reconnectDevice_DISABLED' | Set-Content infrastructure\whatsapp\device_health_monitor.go"

echo.
echo ================================================
echo All Auto-Reconnection Functions Disabled!
echo ================================================
echo.
echo Changes made:
echo - Device Health Monitor will not start
echo - Error Monitor will not monitor device errors
echo - Auto Connection Monitor returns immediately
echo - Multi-Device Auto Reconnect is disabled
echo - Device health check loops are disabled
echo.
echo Devices will now:
echo - Stay in their current state (online/offline)
echo - NOT auto-reconnect when disconnected
echo - NOT be forced to reconnect during campaigns
echo.
pause
