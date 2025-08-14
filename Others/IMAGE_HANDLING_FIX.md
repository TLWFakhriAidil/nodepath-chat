# Campaign & Sequence Image Handling Fix

## Issue Identified
When creating campaigns or sequences with images, the images were not being sent properly.

## Root Causes Found

### 1. **Query Alias Mismatch**
- The query was selecting `media_url AS image_url`
- But scanning into `&msg.MediaURL`
- This caused the MediaURL field to be empty

### 2. **Field Mapping Issues**
- BroadcastMessage struct has both `MediaURL` and `ImageURL` fields
- Device worker uses `msg.MediaURL` to download images
- Frontend sends `image_url` which needs to map to `media_url` in database

## Fixes Applied

### 1. **Broadcast Repository Query**
```sql
-- Before
SELECT ... bm.media_url AS image_url ...

-- After  
SELECT ... bm.media_url ...
```

### 2. **Backward Compatibility**
Added code to set both fields after scanning:
```go
// Set ImageURL for backward compatibility
msg.ImageURL = msg.MediaURL
msg.Message = msg.Content
```

### 3. **Campaign Creation**
- Campaign stores image in `ImageURL` field
- When queueing, it sets `MediaURL = campaign.ImageURL`

### 4. **Sequence Creation**
- Sequence steps store in `media_url` column
- Direct broadcast processor sets both `MediaURL` and `ImageURL`

## Message Flow with Images

1. **Campaign/Sequence Creation**
   - Frontend sends `image_url` field
   - Saved to database `media_url` column

2. **Message Queueing**
   - Campaign: `MediaURL = campaign.ImageURL`
   - Sequence: `MediaURL = step.MediaURL`

3. **Worker Processing**
   - Reads from `broadcast_messages.media_url`
   - Sets both `MediaURL` and `ImageURL` for compatibility

4. **Sending**
   - Device worker uses `msg.MediaURL` to download image
   - Uploads to WhatsApp and sends

## Testing

### Campaign with Image
1. Create campaign with image URL
2. Campaign triggers and queues messages
3. Worker downloads image from `MediaURL`
4. Sends image with caption

### Sequence with Image  
1. Create sequence with image steps
2. Direct broadcast queues with `MediaURL`
3. Worker processes same as campaigns

## Status
✅ Fixed - Images now work properly for both campaigns and sequences!

The key was ensuring the `media_url` column is properly mapped to the `MediaURL` field throughout the entire flow.
