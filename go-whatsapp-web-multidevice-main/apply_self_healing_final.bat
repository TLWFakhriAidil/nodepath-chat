@echo off
echo ========================================
echo APPLYING SELF-HEALING CHANGES
echo ========================================
echo.

:: Check if we're in the right directory
if not exist "src\cmd\rest.go" (
    echo ERROR: Cannot find src\cmd\rest.go
    echo Make sure you run this from the project root directory
    exit /b 1
)

echo Step 1: Creating backups...
copy "src\cmd\rest.go" "src\cmd\rest.go.backup" >nul 2>&1
copy "src\infrastructure\whatsapp\client_manager.go" "src\infrastructure\whatsapp\client_manager.go.backup" >nul 2>&1
echo Backups created!
echo.

echo Step 2: Disabling health monitor in rest.go...
powershell -Command "(Get-Content 'src\cmd\rest.go') -replace 'healthMonitor := whatsapp.GetDeviceHealthMonitor\(whatsappDB\)', '// DISABLED - Using self-healing per message instead`n`t// healthMonitor := whatsapp.GetDeviceHealthMonitor(whatsappDB)' | Set-Content 'src\cmd\rest.go'"
powershell -Command "(Get-Content 'src\cmd\rest.go') -replace 'healthMonitor.Start\(\)', '// healthMonitor.Start()' | Set-Content 'src\cmd\rest.go'"
powershell -Command "(Get-Content 'src\cmd\rest.go') -replace 'Device health monitor started - STATUS CHECK ONLY \(no auto reconnect\)', 'SELF-HEALING MODE: Workers refresh clients per message (no background keepalive)' | Set-Content 'src\cmd\rest.go'"
echo Health monitor disabled!
echo.

echo Step 3: Removing keepalive calls from client_manager.go...
:: This is trickier, let's create a Python script to do it properly
echo import re > temp_fix.py
echo. >> temp_fix.py
echo with open('src/infrastructure/whatsapp/client_manager.go', 'r') as f: >> temp_fix.py
echo     content = f.read() >> temp_fix.py
echo. >> temp_fix.py
echo # Remove keepalive calls in AddClient >> temp_fix.py
echo pattern1 = r'(\s*km := GetKeepaliveManager\(\)\s*\n\s*km\.StartKeepalive\(deviceID, client\)\s*\n)' >> temp_fix.py
echo content = re.sub(pattern1, '\t// DISABLED - Using self-healing instead\n\t// km := GetKeepaliveManager()\n\t// km.StartKeepalive(deviceID, client)\n', content) >> temp_fix.py
echo. >> temp_fix.py
echo # Remove keepalive calls in RemoveClient >> temp_fix.py
echo pattern2 = r'(\s*km := GetKeepaliveManager\(\)\s*\n\s*km\.StopKeepalive\(deviceID\)\s*\n)' >> temp_fix.py
echo content = re.sub(pattern2, '\t// DISABLED - Using self-healing instead\n\t// km := GetKeepaliveManager()\n\t// km.StopKeepalive(deviceID)\n', content) >> temp_fix.py
echo. >> temp_fix.py
echo with open('src/infrastructure/whatsapp/client_manager.go', 'w') as f: >> temp_fix.py
echo     f.write(content) >> temp_fix.py

python temp_fix.py
del temp_fix.py
echo Keepalive calls removed!
echo.

echo Step 4: Verifying changes...
echo.
echo Checking rest.go:
findstr /C:"SELF-HEALING MODE" "src\cmd\rest.go" >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] Self-healing mode message found
) else (
    echo [WARNING] Self-healing mode message not found
)

echo.
echo ========================================
echo SELF-HEALING CHANGES APPLIED!
echo ========================================
echo.
echo Next steps:
echo 1. Review the changes manually if needed
echo 2. Run: build_local.bat
echo 3. Test with: whatsapp.exe rest --db-uri="your-connection-string"
echo 4. Monitor logs for self-healing messages
echo.
echo To revert changes:
echo - Restore from backups: rest.go.backup and client_manager.go.backup
echo.
pause
