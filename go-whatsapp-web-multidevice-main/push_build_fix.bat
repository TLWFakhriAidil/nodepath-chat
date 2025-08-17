@echo off
echo ========================================
echo Pushing Build Fixes
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Creating commit...
git commit -m "Fix: Build errors - remove duplicate GetAllClients method

- Removed duplicate GetAllClients method declaration
- Fixed unused variable 'bm' in rest.go
- Build now completes successfully
- Ready for deployment"

echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Push Complete!
echo ========================================
echo.
echo Build is now fixed and ready for deployment.
echo.
pause
