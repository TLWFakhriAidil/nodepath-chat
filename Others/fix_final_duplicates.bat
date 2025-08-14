@echo off
echo Fixing remaining duplicates...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

REM Create a temporary file with line numbers to remove
echo Creating temporary removal script...

REM Use PowerShell to remove specific line ranges
powershell -Command "$content = Get-Content 'src\ui\rest\app.go'; $newContent = @(); $skip = $false; $skipUntil = 0; for ($i = 0; $i -lt $content.Length; $i++) { $line = $content[$i]; $lineNum = $i + 1; if ($lineNum -eq 1704 -and $line -match 'GetWorkerStatus') { $skip = $true; $skipUntil = 1783 } elseif ($lineNum -eq 1785 -and $line -match 'func min') { $skip = $true; $skipUntil = 1791 } elseif ($lineNum -eq 1793 -and $line -match 'countConnectedDevices') { $skip = $true; $skipUntil = 1800 } elseif ($lineNum -gt $skipUntil) { $skip = $false }; if (-not $skip) { $newContent += $line } }; $newContent | Set-Content 'src\ui\rest\app.go' -Encoding UTF8"

echo Testing build...
cd src
go build .
cd ..

if %ERRORLEVEL% EQU 0 (
    echo Build successful!
    echo Committing and pushing...
    git add -A
    git commit -m "Fix: Remove final duplicate functions - GetWorkerStatus, min, countConnectedDevices"
    git push origin main --force
    echo Done!
) else (
    echo Build still has errors.
)

pause
