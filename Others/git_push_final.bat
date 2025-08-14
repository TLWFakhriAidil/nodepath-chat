@echo off
echo ========================================
echo Git Status Check
echo ========================================
cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

git status

echo.
echo ========================================
echo Adding all changes...
echo ========================================
git add .

echo.
echo ========================================
echo Creating commit...
echo ========================================
git commit -m "Fix critical duplicate and ordering issues for both sequences and campaigns

FIXES APPLIED:
1. Duplicate Prevention for BOTH systems:
   - Sequences: Check sequence_stepid + recipient_phone + device_id
   - Campaigns: Check campaign_id + recipient_phone + device_id
   - Skip insertion if duplicate exists (prevents multiple messages)

2. Message Ordering Fix:
   - Changed ORDER BY from created_at to scheduled_at
   - Ensures messages are sent in correct chronological order
   - Applies to both sequences and campaigns

3. Code Changes:
   - Updated broadcast_repository.go QueueMessage() with duplicate checks
   - Fixed GetPendingMessages() query ordering
   - Built and tested successfully

4. Worker System:
   - Verified mutex locking prevents race conditions
   - Each device has dedicated worker with proper synchronization

This resolves the issues of:
- 132+ duplicate messages being created
- Messages sent in wrong order (Day 3 before Day 1)
- Prevents future duplicates from occurring"

echo.
echo ========================================
echo Pushing to GitHub main branch...
echo ========================================
git push origin main

echo.
echo ========================================
echo Push complete! Check GitHub for updates.
echo ========================================
pause
