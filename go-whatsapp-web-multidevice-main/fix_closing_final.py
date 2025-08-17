# Properly remove everything after the first </script></body></html>
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Find the first occurrence of the closing tags
import re
match = re.search(r'</script>\s*</body>\s*</html>', content)
if match:
    # Cut everything after the closing tags
    content = content[:match.end()]
    
    with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
        f.write(content)
    
    print("Removed all content after closing tags!")
else:
    print("Could not find closing tags pattern")
