# Sequence Backend Fixes Needed

## Issues Found

### 1. **Time Schedule Not Saving**
The frontend is sending `time_schedule` but the backend `CreateSequence` function in `src/usecase/sequence.go` is not saving it to the database.

**Fix needed in `src/usecase/sequence.go`:**
```go
sequence := &models.Sequence{
    UserID:          request.UserID,
    Name:            request.Name,
    Description:     request.Description,
    Niche:           request.Niche,
    TimeSchedule:    request.TimeSchedule,  // ADD THIS LINE
    MinDelaySeconds: request.MinDelaySeconds,  // ADD THIS LINE
    MaxDelaySeconds: request.MaxDelaySeconds,  // ADD THIS LINE
    TotalDays:       len(request.Steps),
    IsActive:        request.IsActive,
    Status:          request.Status,
}
```

### 2. **Sequence Steps Missing Fields**
The step creation is missing several fields that exist in the database schema.

**Fix needed in `src/usecase/sequence.go`:**
```go
step := &models.SequenceStep{
    SequenceID:      sequence.ID,
    Day:             stepReq.Day,
    DayNumber:       stepReq.DayNumber,  // ADD THIS LINE
    MessageType:     stepReq.MessageType,
    Content:         stepReq.Content,
    MediaURL:        stepReq.MediaURL,
    Caption:         stepReq.Caption,
    SendTime:        stepReq.SendTime,  // ADD THIS LINE
    MinDelaySeconds: stepReq.MinDelaySeconds,  // ADD THIS LINE if exists
    MaxDelaySeconds: stepReq.MaxDelaySeconds,  // ADD THIS LINE if exists
}
```

### 3. **Update Sequence Function**
The `UpdateSequence` function also needs to save time_schedule and handle steps properly.

## Frontend Data Being Sent

The frontend is now sending this structure for sequences:
```json
{
    "name": "Sequence Name",
    "description": "Description",
    "niche": "Tag value",
    "time_schedule": "09:00",
    "min_delay_seconds": 10,
    "max_delay_seconds": 30,
    "status": "draft",
    "steps": [
        {
            "day": 1,
            "day_number": 1,
            "content": "Message text",
            "image_url": "base64...",
            "media_url": "base64...",
            "message_type": "text",
            "send_time": "09:00",
            "min_delay_seconds": 10,
            "max_delay_seconds": 30
        }
    ]
}
```

## Database Schema

The `sequences` table has these columns that need to be populated:
- `time_schedule` (VARCHAR)
- `min_delay_seconds` (INT)
- `max_delay_seconds` (INT)

The `sequence_steps` table has:
- `day` (INT)
- `send_time` (VARCHAR)
- `message_type` (VARCHAR)
- `content` (TEXT)
- `media_url` (TEXT)

## Testing

After backend fixes:
1. Create a sequence with time schedule and delays
2. Add day templates
3. Check database to verify all fields are saved
4. View/Edit sequence to verify data loads correctly

## SQL to Verify

```sql
-- Check if sequence has time_schedule
SELECT id, name, time_schedule, min_delay_seconds, max_delay_seconds 
FROM sequences 
ORDER BY created_at DESC;

-- Check if steps are created with all fields
SELECT * FROM sequence_steps 
WHERE sequence_id = (SELECT id FROM sequences ORDER BY created_at DESC LIMIT 1);
```
