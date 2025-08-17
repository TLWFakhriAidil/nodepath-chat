// Fix for Device Status Not Updating After Successful Connection

The issue is that the device pairs successfully but the status remains "offline" in the database.

We need to:
1. Store the current device ID when QR is scanned
2. Update device status to "online" when Connected event fires
3. Save the WhatsApp JID (phone number) to the device

Here's what's happening:
1. QR scan successful ✓
2. WhatsApp shows as linked on phone ✓
3. PairSuccess event fires ✓
4. But device status not updated in database ✗

The fix requires:
- Tracking which device is being connected
- Updating the device record with:
  - status = "online"
  - phone = WhatsApp phone number
  - jid = WhatsApp JID
  - lastSeen = current time
