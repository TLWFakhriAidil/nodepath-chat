@echo off
echo Fixing duplicate StopAllWorkers function...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

REM Create a PowerShell script to remove the duplicate function
echo $content = Get-Content 'src\ui\rest\app.go' -Raw > fix_duplicate.ps1
echo # Remove the first StopAllWorkers function (lines ~1703-1748) >> fix_duplicate.ps1
echo $pattern = '(?s)\n// StopAllWorkers stops all running device workers\nfunc \(handler \*App\) StopAllWorkers\(c \*fiber\.Ctx\) error \{[^}]+for _, device := range devices[^}]+\}[^}]+\}[^}]+\}[\s]*' >> fix_duplicate.ps1
echo $content = $content -replace $pattern, "`n" >> fix_duplicate.ps1
echo Set-Content 'src\ui\rest\app.go' -Value $content -NoNewline >> fix_duplicate.ps1

powershell -ExecutionPolicy Bypass -File fix_duplicate.ps1
del fix_duplicate.ps1

echo Testing build...
cd src
go build .
cd ..

if %ERRORLEVEL% EQU 0 (
    echo Build successful! Committing and pushing...
    git add -A
    git commit -m "Fix: Remove duplicate StopAllWorkers function with unused device variable"
    git push origin main --force
    echo Done!
) else (
    echo Build failed.
)

pause
