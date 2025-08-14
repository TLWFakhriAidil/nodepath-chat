# Fix for Multi-Device Global Client Issue

## Problem
The system uses a global `cli` variable in the whatsapp package, which causes all devices to show the same phone number. When any device connects, it uses this global client instead of the device-specific client.

## Root Cause
1. In `init.go`, there's a global event handler that uses `cli` variable
2. The `handleConnectionEvents` function uses this global `cli`
3. When multiple devices connect, they all reference the same global client

## Solution
We need to modify the event handler to pass the correct client instance to each handler function, instead of using the global `cli` variable.

## Changes Required
1. Modify event handlers to accept the client as a parameter
2. Remove dependency on global `cli` variable
3. Ensure each device uses its own client instance