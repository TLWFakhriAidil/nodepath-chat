@echo off
echo Fixing and pushing build error fixes...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Testing build...
go build -o whatsapp_test.exe src/main.go

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Build successful! Pushing fix...
    
    git add src/infrastructure/whatsapp/auto_connection_monitor_15min.go
    git commit -m "fix: Fix build errors in auto_connection_monitor

- Remove unused whatsmeow import
- Change UserRepositoryInterface to *UserRepository
- Fix repository type definition"
    
    git push origin main
    
    echo.
    echo ✅ Fix pushed to GitHub!
    
    REM Clean up test build
    del whatsapp_test.exe 2>nul
) else (
    echo.
    echo ❌ Build failed. Please check error messages above.
)

pause