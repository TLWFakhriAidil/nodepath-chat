# Multi-Device WhatsApp Connection Fix

## Problem
When scanning QR codes for multiple devices, all devices show the same phone number (60146674397) instead of their respective numbers. This happens because:

1. The system uses a global WhatsApp client instead of device-specific clients
2. Device lookup is done by phone number, which causes conflicts when multiple devices try to connect
3. The connection session mechanism isn't properly linking device IDs with their WhatsApp connections

## Solution Plan

### 1. Store Device ID in WhatsApp Store
We need to pass the device ID to the WhatsApp connection so it knows which device it belongs to.

### 2. Fix Device Registration
Instead of looking up devices by phone number (which can be duplicate), we need to use the device ID from the connection session.

### 3. Create Device-Specific Clients
Each device should have its own WhatsApp client instance, not share a global one.

## Implementation Steps

1. Modify the QR generation to use device-specific clients
2. Store device ID in the WhatsApp session before connecting
3. Fix the device lookup to use session-stored device ID instead of phone number
4. Ensure each device maintains its own WhatsApp connection

This will allow true multi-device support where each device can connect to a different WhatsApp number.