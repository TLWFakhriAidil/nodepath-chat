# Remove unused strings import
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository\broadcast_repository.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Remove the strings import
content = content.replace('\t"strings"\n', '')

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository\broadcast_repository.go', 'w', encoding='utf-8') as f:
    f.write(content)

print("Removed unused strings import")
