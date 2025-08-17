@echo off
echo Fixing unused variable error...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

REM Create a temporary PowerShell script to fix the issue
echo $content = Get-Content 'src\ui\rest\app.go' -Raw > fix_unused_var.ps1
echo $content = $content -replace 'for _, device := range devices \{', 'for range devices {' >> fix_unused_var.ps1
echo Set-Content 'src\ui\rest\app.go' -Value $content >> fix_unused_var.ps1

REM Run the PowerShell script
powershell -ExecutionPolicy Bypass -File fix_unused_var.ps1

REM Clean up
del fix_unused_var.ps1

echo Fixed! Now committing and pushing...

git add -A
git commit -m "Fix: Remove unused device variable in worker control functions"
git push origin main --force

echo Done!
pause
