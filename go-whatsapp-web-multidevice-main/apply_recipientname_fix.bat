@echo off
echo ========================================
echo Adding RecipientName to Sequence Messages
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase"

echo Creating backup...
copy sequence_trigger_processor.go sequence_trigger_processor_backup.go

echo.
echo Applying fix...

REM Use PowerShell to add RecipientName after RecipientPhone
powershell -Command "(Get-Content sequence_trigger_processor.go) | ForEach-Object { if ($_ -match '^\s+RecipientPhone:\s+job\.phone,$') { $_; '		RecipientName:  job.name,' } else { $_ } } | Set-Content sequence_trigger_processor_fixed.go"

REM Replace the original file
move /Y sequence_trigger_processor_fixed.go sequence_trigger_processor.go

echo.
echo Fix applied successfully!
echo RecipientName field added to broadcast message.
cd ..\..
