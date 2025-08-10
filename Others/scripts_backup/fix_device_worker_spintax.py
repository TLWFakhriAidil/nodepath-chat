import re

# Read the device_worker.go file
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\infrastructure\broadcast\device_worker.go', 'r') as f:
    content = f.read()

# Add antipattern import if not present
if '"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/antipattern"' not in content:
    # Find the import block and add the antipattern import
    import_pattern = r'(import \([^)]+)'
    replacement = r'\1\t"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/antipattern"\n'
    content = re.sub(import_pattern, replacement, content, count=1)

# Add fields to DeviceWorker struct
struct_pattern = r'(type DeviceWorker struct \{[^}]+)(broadcastRepo\s+\*repository\.BroadcastRepository)'
replacement = r'\1\2\n\tgreetingProcessor *antipattern.GreetingProcessor\n\tmessageRandomizer *antipattern.MessageRandomizer'
content = re.sub(struct_pattern, replacement, content, flags=re.DOTALL)

# Update NewDeviceWorker to initialize the processors
new_worker_pattern = r'(broadcastRepo:\s+repository\.GetBroadcastRepository\(\),)'
replacement = r'\1\n\t\tgreetingProcessor: antipattern.NewGreetingProcessor(),\n\t\tmessageRandomizer: antipattern.NewMessageRandomizer(),'
content = re.sub(new_worker_pattern, replacement, content)

# Update sendTextMessage function
old_send_text = r'// sendTextMessage sends text message\nfunc \(dw \*DeviceWorker\) sendTextMessage\(recipient types\.JID, msg domainBroadcast\.BroadcastMessage\) error \{\n\tmessage := &waProto\.Message\{\n\t\tExtendedTextMessage: &waProto\.ExtendedTextMessage\{\n\t\t\tText: &msg\.Content,\n\t\t\},\n\t\}\n\t\n\t_, err := dw\.client\.SendMessage\(context\.Background\(\), recipient, message\)\n\treturn err\n\}'

new_send_text = '''// sendTextMessage sends text message
func (dw *DeviceWorker) sendTextMessage(recipient types.JID, msg domainBroadcast.BroadcastMessage) error {
	// Process greeting with spintax
	processedContent := dw.greetingProcessor.PrepareMessageWithGreeting(
		msg.Content,
		msg.RecipientName,
		dw.deviceID,
		msg.RecipientPhone,
	)
	
	// Apply randomization techniques
	finalContent := dw.messageRandomizer.RandomizeMessage(processedContent)
	
	message := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: &finalContent,
		},
	}
	
	_, err := dw.client.SendMessage(context.Background(), recipient, message)
	return err
}'''

content = re.sub(old_send_text, new_send_text, content, flags=re.DOTALL)

# Write the updated content
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\infrastructure\broadcast\device_worker.go', 'w') as f:
    f.write(content)

print("Successfully added spintax and greeting processing to device_worker.go!")
