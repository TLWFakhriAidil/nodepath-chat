import re

# Read the file
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Fix all occurrences of "else" on its own line followed by a comment or code
# Pattern: line with just "else" (possibly with whitespace) followed by a new line
pattern = r'(\n\s*)else(\s*\n)'
replacement = r'\1} else {\2'

# Apply the fix
fixed_content = re.sub(pattern, replacement, content)

# Write back
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device.html', 'w', encoding='utf-8') as f:
    f.write(fixed_content)

print("Fixed all else statements!")
