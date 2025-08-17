import re

# Read the file
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Look for the section with the campaigns processing
# Find the forEach loop and the misplaced else
pattern = r'(campaignArray\.forEach\(campaign => \{[\s\S]*?\}\);[\s\S]*?renderCalendar\(\);[\s\S]*?)(\} else \{)'

# Check if campaignArray has items before the forEach
replacement = r'\1\n                    if (campaignArray.length > 0) {\n                        renderCalendar();\n                    \2'

# First, let's find and fix the structure around line 4605
# We need to add the missing if condition
lines = content.split('\n')

# Find the problematic section
for i in range(len(lines)):
    if i > 4600 and i < 4610:
        if 'renderCalendar();' in lines[i]:
            # Add check for campaign array before rendering
            if i+1 < len(lines) and '} else {' in lines[i+1]:
                # Insert the if condition
                lines[i] = '                    if (campaignArray && campaignArray.length > 0) {\n' + lines[i]

# Join back
fixed_content = '\n'.join(lines)

# Write back
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device.html', 'w', encoding='utf-8') as f:
    f.write(fixed_content)

print("Fixed the else statement structure!")
