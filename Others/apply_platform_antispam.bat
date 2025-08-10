@echo off
echo ========================================
echo Adding Anti-Spam to Platform Messages
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"

echo Backing up files...
copy "pkg\external\platform_sender.go" "pkg\external\platform_sender.go.backup_before_antispam" >nul
copy "infrastructure\broadcast\whatsapp_message_sender.go" "infrastructure\broadcast\whatsapp_message_sender.go.backup_before_antispam" >nul

echo.
echo Applying anti-spam to platform messages...

REM Update WhatsApp message sender to pass all needed data to platform sender
powershell -Command @"
$file = 'infrastructure\broadcast\whatsapp_message_sender.go'
$content = Get-Content $file -Raw

# Update the platform sender call to include recipient name and device ID
$content = $content -replace '(err = w\.platformSender\.SendMessage\()\s*device\.Platform,\s*instance,\s*msg\.RecipientPhone,\s*msg\.Message,\s*msg\.ImageURL,\s*\)', '$1device.Platform, instance, msg.RecipientPhone, msg.RecipientName, msg.Message, msg.ImageURL, deviceID)'

Set-Content $file $content
Write-Host 'Updated WhatsApp message sender'
"@

REM Update platform sender interface
powershell -Command @"
$file = 'pkg\external\platform_sender.go'
$content = Get-Content $file -Raw

# Update SendMessage signature
$content = $content -replace 'func \(ps \*PlatformSender\) SendMessage\(platform, instance, phone, message, imageURL string\)', 'func (ps *PlatformSender) SendMessage(platform, instance, phone, recipientName, message, imageURL, deviceID string)'

# Update the switch case calls
$content = $content -replace 'return ps\.sendViaWablas\(instance, phone, message, imageURL\)', 'return ps.sendViaWablas(instance, phone, recipientName, message, imageURL, deviceID)'
$content = $content -replace 'return ps\.sendViaWhacenter\(instance, phone, message, imageURL\)', 'return ps.sendViaWhacenter(instance, phone, recipientName, message, imageURL, deviceID)'

# Update method signatures
$content = $content -replace 'func \(ps \*PlatformSender\) sendViaWablas\(token, phone, message, imageURL string\)', 'func (ps *PlatformSender) sendViaWablas(token, phone, recipientName, message, imageURL, deviceID string)'
$content = $content -replace 'func \(ps \*PlatformSender\) sendViaWhacenter\(instance, phone, message, imageURL string\)', 'func (ps *PlatformSender) sendViaWhacenter(instance, phone, recipientName, message, imageURL, deviceID string)'

# Add anti-spam processing method
$antiSpamMethod = @'

// applyAntiSpam applies greeting and randomization to message
func (ps *PlatformSender) applyAntiSpam(message, recipientName, deviceID, phone string) string {
	// Add Malaysian greeting
	messageWithGreeting := ps.greetingProcessor.PrepareMessageWithGreeting(
		message,
		recipientName,
		deviceID,
		phone,
	)
	
	// Apply randomization
	randomizedMessage := ps.messageRandomizer.RandomizeMessage(messageWithGreeting)
	
	logrus.Debugf("Platform anti-spam applied: %d transformations", strings.Count(randomizedMessage, message))
	
	return randomizedMessage
}
'@

# Insert before the last closing brace
$lastBrace = $content.LastIndexOf('}')
$content = $content.Insert($lastBrace, $antiSpamMethod)

# Update sendWablasText to apply anti-spam
$content = $content -replace '(data\.Set\("message", )message\)', '$1ps.applyAntiSpam(message, recipientName, deviceID, phone))'

# Update sendWablasImage caption
$content = $content -replace '(data\.Set\("caption", )caption\)', '$1ps.applyAntiSpam(caption, recipientName, deviceID, phone))'

# Update method calls in sendViaWablas
$content = $content -replace 'return ps\.sendWablasImage\(token, phone, message, imageURL\)', 'return ps.sendWablasImage(token, phone, recipientName, message, imageURL, deviceID)'
$content = $content -replace 'return ps\.sendWablasText\(token, phone, message\)', 'return ps.sendWablasText(token, phone, recipientName, message, deviceID)'

# Update sendWablasText signature
$content = $content -replace 'func \(ps \*PlatformSender\) sendWablasText\(token, phone, message string\)', 'func (ps *PlatformSender) sendWablasText(token, phone, recipientName, message, deviceID string)'

# Update sendWablasImage signature  
$content = $content -replace 'func \(ps \*PlatformSender\) sendWablasImage\(token, phone, caption, imageURL string\)', 'func (ps *PlatformSender) sendWablasImage(token, phone, recipientName, caption, imageURL, deviceID string)'

Set-Content $file $content
Write-Host 'Updated platform sender with anti-spam'
"@

echo.
echo Done! Platform messages (Wablas/Whacenter) now have anti-spam features:
echo - Malaysian greetings
echo - Message randomization (homoglyphs, zero-width chars)
echo - Same anti-spam as WhatsApp Web messages
echo.
echo Building to test...
set CGO_ENABLED=0
go build -o ../whatsapp.exe .
echo.
echo Build complete!
pause
