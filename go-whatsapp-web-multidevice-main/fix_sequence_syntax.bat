@echo off
echo Fixing sequence syntax error...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Creating backup...
copy src\views\dashboard.html src\views\dashboard_backup_syntax_fix.html

echo Fixing file...
powershell -Command "(Get-Content src\views\dashboard.html) | ForEach-Object { if ($_ -match '^\s*// Campaign Functions\s*$' -and $LastLine -match 'Swal.fire') { $_ } elseif ($_ -notmatch '<input type=\"hidden\" class=\"step-image-url\">' -and $_ -notmatch '<div class=\"step-image-preview' -and $_ -notmatch 'Min Delay \(seconds\)' -and $_ -notmatch 'Max Delay \(seconds\)' -and $_ -notmatch 'stepsContainer\.insertAdjacentHTML' -and $_ -notmatch 'function removeStep' -and $_ -notmatch 'function updateStepNumbers' -and $_ -notmatch 'function compressStepImage' -and $_ -notmatch 'Add event listener for Add Step button' -and $_ -notmatch 'addStepBtn\.addEventListener') { $_ }; $LastLine = $_ } | Set-Content src\views\dashboard_temp.html"

move /Y src\views\dashboard_temp.html src\views\dashboard.html

echo Done!
pause
