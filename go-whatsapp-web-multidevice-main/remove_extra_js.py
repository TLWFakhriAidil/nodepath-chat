import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    lines = f.readlines()

# Find the first proper closing
closing_line = -1
for i, line in enumerate(lines):
    if '</script>' in line and i > 500:  # After main script content
        # Check if next lines have </body> and </html>
        if i+1 < len(lines) and '</body>' in lines[i+1]:
            if i+2 < len(lines) and '</html>' in lines[i+2]:
                closing_line = i + 3
                break

if closing_line > 0:
    # Keep only up to the proper closing
    lines = lines[:closing_line]
    
    # Write back
    with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
        f.writelines(lines)
    
    print(f"Removed {len(lines) - closing_line} lines of JavaScript after closing tags!")
else:
    print("Could not find proper closing tags")
