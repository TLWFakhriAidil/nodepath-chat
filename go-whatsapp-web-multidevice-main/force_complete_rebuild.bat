@echo off
echo Cleaning up old files...
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

REM Delete the embedded views to force complete rebuild
echo Temporarily moving views folder...
move src\views src\views_backup

REM Commit this state
git add -A
git commit -m "Temporarily remove views folder to force complete rebuild"
git push origin main

REM Wait a moment
timeout /t 5

REM Restore the views
echo Restoring views folder...
move src\views_backup src\views

REM Commit the restored state
git add -A
git commit -m "Restore views folder with all fixes - force Railway to see changes"
git push origin main

echo.
echo ============================================
echo FORCED COMPLETE REBUILD
echo ============================================
echo.
echo This will force Railway to:
echo 1. See that views folder is gone (invalidate ALL caches)
echo 2. Then see views folder is back with fixes
echo 3. Force complete recompilation of embedded files
echo.
pause