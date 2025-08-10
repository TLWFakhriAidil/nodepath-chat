@echo off
echo Applying anti-pattern fix to sequences...

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase"

REM Create a backup
copy sequence_trigger_processor.go sequence_trigger_processor_backup.go

REM Apply the fix using PowerShell
powershell -Command "(Get-Content sequence_trigger_processor.go) -replace 'broadcastMsg := domainBroadcast.BroadcastMessage\{', 'broadcastMsg := domainBroadcast.BroadcastMessage{' | Set-Content sequence_trigger_processor_temp.go"

REM Now add RecipientName
powershell -Command "(Get-Content sequence_trigger_processor_temp.go) -replace '(\s+RecipientPhone:\s+job\.phone,)', '$1`r`n`t`tRecipientName:  job.name,' | Set-Content sequence_trigger_processor_temp2.go"

REM Add min/max delays
powershell -Command "(Get-Content sequence_trigger_processor_temp2.go) -replace '(\s+Type:\s+job\.messageType,)', '$1`r`n`t`tMinDelay:       job.minDelaySeconds,`r`n`t`tMaxDelay:       job.maxDelaySeconds,' | Set-Content sequence_trigger_processor_fixed.go"

REM Replace original
move /Y sequence_trigger_processor_fixed.go sequence_trigger_processor.go

REM Clean up temps
del sequence_trigger_processor_temp.go
del sequence_trigger_processor_temp2.go

echo Fix applied!
cd ..\..
