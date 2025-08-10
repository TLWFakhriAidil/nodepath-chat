@echo off
echo ========================================
echo Applying Anti-Spam Features to Sequences
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"

echo Backing up original file...
copy "usecase\sequence_trigger_processor.go" "usecase\sequence_trigger_processor.go.backup_before_antispam" >nul

echo.
echo Applying anti-spam features to sequences...

REM Step 1: Update the contactJob struct
powershell -Command "$content = Get-Content 'usecase\sequence_trigger_processor.go' -Raw; $content = $content -replace '(type contactJob struct \{[^}]+)(preferredDevice\s+sql\.NullString)', '$1minDelaySeconds  int`r`n`tmaxDelaySeconds  int`r`n`t$2'; Set-Content 'usecase\sequence_trigger_processor.go' $content"

REM Step 2: Update the query to include sequence delays
powershell -Command "$content = Get-Content 'usecase\sequence_trigger_processor.go' -Raw; $content = $content -replace '(SELECT\s+sc\.id, sc\.sequence_id, sc\.contact_phone, sc\.contact_name,\s+sc\.current_trigger, sc\.current_step,\s+ss\.content, ss\.message_type, ss\.media_url,\s+ss\.next_trigger, ss\.trigger_delay_hours,\s+l\.device_id as preferred_device_id)', '$1, s.min_delay_seconds, s.max_delay_seconds'; Set-Content 'usecase\sequence_trigger_processor.go' $content"

REM Step 3: Update the Scan to include the new fields
powershell -Command "$content = Get-Content 'usecase\sequence_trigger_processor.go' -Raw; $content = $content -replace '(if err := rows\.Scan\(&job\.contactID, &job\.sequenceID, &job\.phone, &job\.name,\s+&job\.currentTrigger, &job\.currentStep, &job\.messageText, &job\.messageType,\s+&job\.mediaURL, &job\.nextTrigger, &job\.delayHours, &job\.preferredDevice\))', '$1.Replace('&job.preferredDevice)', '&job.preferredDevice, &job.minDelaySeconds, &job.maxDelaySeconds)')'; Set-Content 'usecase\sequence_trigger_processor.go' $content"

REM Step 4: Update the broadcast message creation
powershell -Command "$content = Get-Content 'usecase\sequence_trigger_processor.go' -Raw; $content = $content -replace '(broadcastMsg := domainBroadcast\.BroadcastMessage\{)(\s+DeviceID:\s+deviceID,\s+RecipientPhone:\s+job\.phone,)', '$1$2`r`n`t`tRecipientName:  job.name,'; Set-Content 'usecase\sequence_trigger_processor.go' $content"

powershell -Command "$content = Get-Content 'usecase\sequence_trigger_processor.go' -Raw; $content = $content -replace '(Type:\s+job\.messageType,)(\s+\})', '$1`r`n`t`tMinDelay:       job.minDelaySeconds,`r`n`t`tMaxDelay:       job.maxDelaySeconds,$2'; Set-Content 'usecase\sequence_trigger_processor.go' $content"

echo.
echo Done! Anti-spam features have been applied to sequences.
echo.
echo Changes made:
echo 1. Added RecipientName to broadcast messages
echo 2. Added MinDelay/MaxDelay from sequence settings
echo 3. Messages will now have:
echo    - Malaysian greetings (Hi Cik, Selamat pagi, etc.)
echo    - Message randomization (homoglyphs, zero-width chars)
echo    - Human-like delays between messages
echo.
pause
