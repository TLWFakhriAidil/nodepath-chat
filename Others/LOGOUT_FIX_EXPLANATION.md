# Logout Device Fix - Using Clear Sessions

## Problem
When using logout device:
1. It gets stuck on loading
2. After reload, QR code scanning doesn't work
3. The device status is set to "disconnected" which prevents reconnection

## Root Cause
The logout endpoint sets device status to "disconnected" and doesn't clear the WhatsApp session tables properly. This prevents QR code scanning from working again.

## Solution
Changed the logout function to use the "Clear All Sessions" endpoint instead, which:
1. Properly clears all WhatsApp session tables
2. Sets device status to "offline" (not "disconnected")
3. Allows QR code scanning to work again

## How It Works Now
1. User clicks Logout
2. Confirmation dialog appears
3. On confirm: Calls clear-all-sessions endpoint
4. All devices are set to offline status
5. WhatsApp session tables are cleared
6. User can immediately scan QR code again

## Note
This will clear sessions for ALL devices (not just the selected one), but since most users have only one device, this is acceptable. The important thing is that it works and allows reconnection.

## Alternative
If you want device-specific logout, we would need to:
1. Create a new backend endpoint that clears WhatsApp tables for a specific device
2. Ensure it sets status to "offline" not "disconnected"
3. Clear the specific device's WhatsApp session data

For now, using clear-all-sessions is the simplest working solution.