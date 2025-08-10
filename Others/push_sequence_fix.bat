@echo off
echo === Pushing Sequence Fix to GitHub ===
echo.

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Checking git status...
git status

echo.
echo Adding modified files...
git add src/usecase/direct_broadcast_processor.go

echo.
echo Creating commit...
git commit -m "Fix sequence enrollment MySQL syntax error - add backticks around trigger keyword"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo === Push Complete ===
pause
