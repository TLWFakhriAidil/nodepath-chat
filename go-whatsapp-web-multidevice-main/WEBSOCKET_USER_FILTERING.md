# WebSocket User Filtering Fix

## Problem
When one user scans a QR code, all connected users see the QR popup modal, even for devices they don't own.

## Solution
Implemented user-specific WebSocket broadcasting:

1. **WebSocket Client Tracking**
   - Each WebSocket connection now stores the user ID
   - Messages can be targeted to specific users or devices

2. **Targeted Broadcasting**
   - QR code events include `targetUserId` and `targetDeviceId`
   - Only the device owner receives QR-related popups

3. **Frontend Filtering**
   - Frontend checks if incoming messages are for the current user
   - Ignores messages targeted at other users

## Benefits
- Users only see QR popups for their own devices
- No interference between multiple users
- Better privacy and user experience
- Supports team collaboration without confusion

## Technical Details
- WebSocket connections store user context
- Broadcast messages can specify target user/device
- Frontend filters messages before processing
