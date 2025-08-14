# Railway Auto-Deploy Trigger

Timestamp: July 02, 2025 - 01:10 AM
Purpose: Trigger Railway auto-deployment

## Latest Changes:
- Fixed transfer API endpoint URL and added JSON request body
- Added 'active' status check for device transfer button visibility  
- Fixed AI campaign leads showing as Unknown by checking ai column
- AI campaign leads now properly show names from leads_ai table
- Transfer functionality now works correctly for AI campaigns

## Fixes Applied:
1. Transfer button now sends proper JSON body with device_id and campaign_id
2. Device status 'active' is now recognized for transfer functionality
3. AI leads display actual names instead of "Unknown"

## To Verify Deployment:
1. Check device report - AI campaign leads should show proper names
2. Transfer button should be visible for active/connected devices
3. Transfer functionality should work without "Invalid request body" error

Commit ID: TRIGGER-20250702-FIXES