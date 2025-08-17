import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Remove duplicate closing tags
content = re.sub(r'(</script>\s*</body>\s*</html>\s*)+$', '</script>\n</body>\n</html>', content)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
    f.write(content)

print("Cleaned up closing tags!")
