# Fix by removing the old function starting at line 533

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository\broadcast_repository.go', 'r', encoding='utf-8') as f:
    lines = f.readlines()

# Find the second occurrence and remove everything after line 532
# Keep only up to line 532 (0-indexed, so 533 lines total)
lines = lines[:533]

# Write back
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository\broadcast_repository.go', 'w', encoding='utf-8') as f:
    f.writelines(lines)

print("Removed old GetPendingMessagesAndLock function")
