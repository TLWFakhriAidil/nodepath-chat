# Direct Broadcast UUID Fix

## The Problem
The error "pq: invalid input syntax for type uuid: """ occurs because:
1. PostgreSQL UUID columns cannot accept empty strings
2. The repository's QueueMessage checks if UserID != "" but this passes if UserID contains an empty string
3. When it inserts, PostgreSQL rejects the empty string for UUID type

## The Solution
The fix is already implemented in direct_broadcast_processor.go:
- We validate leads have non-empty device_id and user_id before enrollment
- We don't set SequenceStepID (which was causing issues)
- We use the repository method which handles NULLs

## IMPORTANT: Deployment Steps

1. **Build the application locally:**
```bash
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
build_local.bat
```

2. **Deploy to your server:**
- Copy the new whatsapp.exe to your server
- Restart the service
- The error should stop occurring

## Verification
After deployment, check logs for:
- "✅ Direct enrollment successful for XXX - Created X messages"
- No more UUID errors

## Alternative Quick Fix
If you can't deploy immediately, you can fix the data:
```sql
-- Update any leads with empty UUID strings to NULL
UPDATE leads 
SET device_id = NULL 
WHERE device_id = '';

UPDATE leads 
SET user_id = NULL 
WHERE user_id = '';
```

The code is already fixed in the repository. You just need to deploy the latest build!
