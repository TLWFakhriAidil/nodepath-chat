import re

# Read the device_worker.go file
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\infrastructure\broadcast\device_worker.go', 'r') as f:
    content = f.read()

# Update sendImageMessage function to process caption with spintax
old_caption_line = r'Caption: &msg\.Caption,'

new_caption_block = '''Caption: func() *string {
			if msg.Caption != "" {
				// Process caption with greeting and randomization
				processedCaption := dw.greetingProcessor.PrepareMessageWithGreeting(
					msg.Caption,
					msg.RecipientName,
					dw.deviceID,
					msg.RecipientPhone,
				)
				finalCaption := dw.messageRandomizer.RandomizeMessage(processedCaption)
				return &finalCaption
			}
			return &msg.Caption
		}(),'''

content = re.sub(old_caption_line, new_caption_block, content)

# Write the updated content
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\infrastructure\broadcast\device_worker.go', 'w') as f:
    f.write(content)

print("Successfully added spintax processing to image captions!")
