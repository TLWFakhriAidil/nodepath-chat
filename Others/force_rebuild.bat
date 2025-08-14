@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Force Railway rebuild - update Dockerfile and railway.toml to bust cache"
git push origin main
echo.
echo ============================================
echo IMPORTANT: Railway cache bust initiated!
echo ============================================
echo.
echo Changes made to force rebuild:
echo 1. Updated Dockerfile timestamp
echo 2. Added cache bust comment to railway.toml
echo 3. Added debug step to verify files
echo.
echo This should force Railway to rebuild from scratch.
echo.
pause