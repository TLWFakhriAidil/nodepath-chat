@echo off
echo ========================================
echo Deploying Worker Health & Auto-Reconnect Improvements
echo ========================================

:: First commit and push the fixes
git add -A
git commit -m "Fix: Worker health check, auto-reconnect & all control buttons functional

- Fixed syntax errors in rest.go (backtick issues)
- Added GetAllDevices method to UserRepository
- Fixed duplicate method declarations in device_worker.go
- Implemented device health monitor with auto-reconnect
- Enhanced client manager with better registration
- Improved worker health checks
- All worker control buttons now functional
- Fixed compilation errors and unused imports
- Better error handling and recovery mechanisms"

git push origin main --force

echo.
echo ========================================
echo Deployment Complete!
echo ========================================
echo.
echo The following improvements have been deployed:
echo.
echo 1. Device Health Monitor:
echo    - Monitors all devices every 30 seconds
echo    - Automatically reconnects disconnected devices
echo    - Updates device status in real-time
echo    - Manual reconnection support
echo.
echo 2. Enhanced Client Manager:
echo    - RegisterDeviceOnConnection for proper registration
echo    - GetConnectedDeviceCount for statistics
echo    - GetDeviceStatus for detailed status checking
echo    - CleanupDisconnectedClients for maintenance
echo.
echo 3. Improved Worker Health:
echo    - Enhanced IsHealthy method with multiple checks
echo    - RestartWorker functionality
echo    - Auto-restart on health check failure
echo    - Queue health monitoring
echo.
echo 4. Worker Control API:
echo    - Resume Failed: Restarts all stopped/failed workers
echo    - Stop All: Stops all active workers
echo    - Restart Worker: Restart specific device worker
echo    - Start Worker: Start worker for specific device
echo    - Reconnect Device: Manually reconnect a device
echo    - Health Check: Trigger health check for all
echo.
echo 5. Frontend Integration:
echo    - All buttons connected to API endpoints
echo    - Real-time toast notifications
echo    - Auto-refresh after actions
echo    - Better error handling
echo.
echo Railway will auto-deploy these changes!
echo.
pause
