import re

# Read the file
with open('src/usecase/optimized_broadcast_processor.go', 'r') as f:
    content = f.read()

# Replace GetPendingMessages with GetPendingMessagesAndLock
content = content.replace(
    'messages, err := p.broadcastRepo.GetPendingMessages(deviceID, MESSAGE_BATCH_SIZE)',
    'messages, err := p.broadcastRepo.GetPendingMessagesAndLock(deviceID, MESSAGE_BATCH_SIZE)'
)

# Write back
with open('src/usecase/optimized_broadcast_processor.go', 'w') as f:
    f.write(content)

print("Fixed: Changed GetPendingMessages to GetPendingMessagesAndLock")

# Also check broadcast_worker_processor if it exists
try:
    with open('src/usecase/broadcast_worker_processor.go', 'r') as f:
        content2 = f.read()
    
    if 'GetPendingMessages(' in content2 and 'GetPendingMessagesAndLock' not in content2:
        content2 = content2.replace(
            'broadcastRepo.GetPendingMessages(deviceID,',
            'broadcastRepo.GetPendingMessagesAndLock(deviceID,'
        )
        
        with open('src/usecase/broadcast_worker_processor.go', 'w') as f:
            f.write(content2)
        
        print("Also fixed broadcast_worker_processor.go")
except:
    print("broadcast_worker_processor.go not found or already fixed")
