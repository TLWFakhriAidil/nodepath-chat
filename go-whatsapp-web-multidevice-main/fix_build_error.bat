@echo off
echo Fixing build errors in auto_connection_monitor_15min.go...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Fixing imports and interface issues...

REM Fix the imports and interface
powershell -Command "$content = Get-Content 'src\infrastructure\whatsapp\auto_connection_monitor_15min.go'; $content = $content -replace 'repository.UserRepositoryInterface', '*repository.UserRepository'; $content = $content -replace 'import \(', 'import ('; $content | Set-Content 'src\infrastructure\whatsapp\auto_connection_monitor_15min.go'"

REM Remove unused import if whatsmeow is not used
powershell -Command "$content = Get-Content 'src\infrastructure\whatsapp\auto_connection_monitor_15min.go'; $hasWhatsmeowUsage = $content | Select-String -Pattern 'whatsmeow\.' -Quiet; if (-not $hasWhatsmeowUsage) { $content = $content -replace '.*go.mau.fi/whatsmeow.*\n', '' }; $content | Set-Content 'src\infrastructure\whatsapp\auto_connection_monitor_15min.go'"

echo.
echo Building to test...
go build -o whatsapp.exe src/main.go

if %ERRORLEVEL% EQU 0 (
    echo.
    echo Build successful! Committing and pushing fix...
    
    git add src/infrastructure/whatsapp/auto_connection_monitor_15min.go
    git commit -m "fix: Fix build errors in auto_connection_monitor_15min.go

- Remove unused whatsmeow import
- Fix repository interface type to use pointer
- Ensure proper type definitions"
    
    git push origin main
    
    echo.
    echo ✅ Fix pushed successfully!
) else (
    echo.
    echo ❌ Build still failing. Let me check the exact issue...
)

pause