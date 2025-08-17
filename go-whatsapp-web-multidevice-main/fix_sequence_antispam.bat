@echo off
echo ========================================
echo Applying Anti-Spam Features to Sequences
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

echo 1. Backing up sequence trigger processor...
copy "src\usecase\sequence_trigger_processor.go" "src\usecase\sequence_trigger_processor.go.backup_antispam"

echo.
echo 2. Updating sequence trigger processor to include anti-spam features...

REM Create the PowerShell script to update the file
powershell -Command @"
$file = 'src\usecase\sequence_trigger_processor.go'
$content = Get-Content $file -Raw

# Find the broadcast message creation section and update it
$oldPattern = '// Queue message to broadcast system
\s+// Create broadcast message
\s+broadcastMsg := domainBroadcast\.BroadcastMessage\{
\s+DeviceID:\s+deviceID,
\s+RecipientPhone:\s+job\.phone,
\s+Message:\s+job\.messageText,
\s+Content:\s+job\.messageText,
\s+Type:\s+job\.messageType,
\s+\}'

$newCode = @'
	// Queue message to broadcast system
	// Create broadcast message with anti-spam features
	broadcastMsg := domainBroadcast.BroadcastMessage{
		DeviceID:       deviceID,
		RecipientPhone: job.phone,
		RecipientName:  job.name,  // Added for greeting processor
		Message:        job.messageText,
		Content:        job.messageText,
		Type:           job.messageType,
		// Add delays for human-like behavior (will be randomized by device worker)
		MinDelay:       5,   // Default 5-15 seconds between messages
		MaxDelay:       15,  // Can be overridden by sequence settings
	}
'@

$content = $content -replace $oldPattern, $newCode

# Also need to get sequence delays - add after contactJob query
$getDelaysPattern = 'FROM sequence_contacts sc'
$getDelaysNew = @'
FROM sequence_contacts sc
		INNER JOIN sequences s ON s.id = sc.sequence_id
'@
$content = $content -replace $getDelaysPattern, $getDelaysNew

# Update the SELECT to include sequence delays
$selectPattern = 'SELECT sc\.id, sc\.sequence_id, sc\.contact_phone, sc\.contact_name'
$selectNew = 'SELECT sc.id, sc.sequence_id, sc.contact_phone, sc.contact_name, s.min_delay_seconds, s.max_delay_seconds'
$content = $content -replace $selectPattern, $selectNew

# Update the job struct to include delays
$jobStructPattern = 'type contactJob struct \{'
$jobStructNew = @'
type contactJob struct {
	minDelaySeconds  int
	maxDelaySeconds  int
'@
$content = $content -replace $jobStructPattern, $jobStructNew

# Update the Scan to include delays
$scanPattern = 'if err := rows\.Scan\(&job\.contactID, &job\.sequenceID, &job\.phone, &job\.name,'
$scanNew = 'if err := rows.Scan(&job.contactID, &job.sequenceID, &job.phone, &job.name, &job.minDelaySeconds, &job.maxDelaySeconds,'
$content = $content -replace $scanPattern, $scanNew

# Apply sequence delays to broadcast message
$applyDelaysPattern = 'MinDelay:\s+5,\s+// Default 5-15 seconds between messages
\s+MaxDelay:\s+15,\s+// Can be overridden by sequence settings'
$applyDelaysNew = @'
MinDelay:       job.minDelaySeconds,  // Use sequence-specific delays
		MaxDelay:       job.maxDelaySeconds,  // For human-like behavior
'@
$content = $content -replace $applyDelaysPattern, $applyDelaysNew

Set-Content $file $content
Write-Host 'Updated sequence trigger processor with anti-spam features'
"@

echo.
echo 3. Verifying the WhatsApp message sender already has anti-spam...
powershell -Command "(Get-Content 'src\infrastructure\broadcast\whatsapp_message_sender.go' -Raw) -match 'greetingProcessor|messageRandomizer'"

echo.
echo 4. Creating a documentation file for the changes...

echo # Sequence Anti-Spam Features Applied > SEQUENCE_ANTISPAM_UPDATE.md
echo. >> SEQUENCE_ANTISPAM_UPDATE.md
echo ## Changes Made: >> SEQUENCE_ANTISPAM_UPDATE.md
echo. >> SEQUENCE_ANTISPAM_UPDATE.md
echo 1. **Added RecipientName** to broadcast messages in sequences >> SEQUENCE_ANTISPAM_UPDATE.md
echo    - Now greeting processor can add personalized Malaysian greetings >> SEQUENCE_ANTISPAM_UPDATE.md
echo. >> SEQUENCE_ANTISPAM_UPDATE.md
echo 2. **Added MinDelay/MaxDelay** from sequence settings >> SEQUENCE_ANTISPAM_UPDATE.md
echo    - Each sequence can have custom delays for human-like behavior >> SEQUENCE_ANTISPAM_UPDATE.md
echo    - Prevents pattern detection by WhatsApp >> SEQUENCE_ANTISPAM_UPDATE.md
echo. >> SEQUENCE_ANTISPAM_UPDATE.md
echo 3. **Message Processing Flow**: >> SEQUENCE_ANTISPAM_UPDATE.md
echo    - Sequence creates message with recipient name and delays >> SEQUENCE_ANTISPAM_UPDATE.md
echo    - Broadcast manager queues the message >> SEQUENCE_ANTISPAM_UPDATE.md
echo    - Device worker applies random delay between min/max >> SEQUENCE_ANTISPAM_UPDATE.md
echo    - WhatsApp sender applies greeting + randomization >> SEQUENCE_ANTISPAM_UPDATE.md
echo. >> SEQUENCE_ANTISPAM_UPDATE.md
echo ## Anti-Spam Features Now Active: >> SEQUENCE_ANTISPAM_UPDATE.md
echo. >> SEQUENCE_ANTISPAM_UPDATE.md
echo ### 1. Malaysian Greeting System >> SEQUENCE_ANTISPAM_UPDATE.md
echo - Original: "Special promotion..." >> SEQUENCE_ANTISPAM_UPDATE.md
echo - With Greeting: "Hi Cik, apa khabar\n\nSpecial promotion..." >> SEQUENCE_ANTISPAM_UPDATE.md
echo. >> SEQUENCE_ANTISPAM_UPDATE.md
echo ### 2. Message Randomization >> SEQUENCE_ANTISPAM_UPDATE.md
echo - Homoglyphs substitution >> SEQUENCE_ANTISPAM_UPDATE.md
echo - Zero-width characters >> SEQUENCE_ANTISPAM_UPDATE.md
echo - Random punctuation >> SEQUENCE_ANTISPAM_UPDATE.md
echo - Case variations >> SEQUENCE_ANTISPAM_UPDATE.md
echo. >> SEQUENCE_ANTISPAM_UPDATE.md
echo ### 3. Human-like Delays >> SEQUENCE_ANTISPAM_UPDATE.md
echo - Random delay between sequence min/max seconds >> SEQUENCE_ANTISPAM_UPDATE.md
echo - Typing simulation (without presence) >> SEQUENCE_ANTISPAM_UPDATE.md
echo - No pattern detection >> SEQUENCE_ANTISPAM_UPDATE.md

echo.
echo ========================================
echo Anti-Spam Features Applied to Sequences!
echo ========================================
echo.
echo Next steps:
echo 1. Review the changes in sequence_trigger_processor.go
echo 2. Build the application: build_local.bat
echo 3. Test with a sequence to see the anti-spam in action
echo.
pause
