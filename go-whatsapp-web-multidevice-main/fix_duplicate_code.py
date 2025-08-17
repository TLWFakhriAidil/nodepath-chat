import re

# Read the file
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Find and fix the duplicated code block
# Look for the pattern where we have duplicated code
pattern = r'let date = campaign\.campaign_date; \{[\s\S]*?let date = campaign\.campaign_date;'

# Replace with just the correct version
replacement = 'let date = campaign.campaign_date;'

# Apply the fix
fixed_content = re.sub(pattern, replacement, content)

# Write back
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device.html', 'w', encoding='utf-8') as f:
    f.write(fixed_content)

print("Fixed the duplicated code block!")
