# Sequence Contacts Table Structure

Based on the actual database, the `sequence_contacts` table has these columns:

## Columns:
- `id` - UUID primary key
- `sequence_id` - References sequences table
- `contact_phone` - Phone number of contact
- `contact_name` - Name of contact
- `current_step` - Current step number in sequence (NOT current_day)
- `status` - Status (active, completed, paused)
- `completed_at` - Timestamp (used for both enrollment date and completion)
- `current_trigger` - Current trigger being processed
- `next_trigger_time` - When to process next trigger
- `processing_device_id` - Device currently processing
- `last_error` - Last error message if any
- `retry_count` - Number of retries
- `assigned_device_id` - Assigned device
- `processing_started_at` - When processing started
- `sequence_stepid` - Reference to specific sequence step

## Important Notes:
1. NO `enrolled_at` column - use `completed_at` for enrollment tracking
2. NO `added_at` column - use `completed_at`
3. NO `current_day` column - use `current_step`
4. NO `last_message_at` column in actual table
5. NO `last_sent_at` column in actual table

## Code Changes Made:
1. All INSERT queries now use `completed_at` for initial enrollment
2. All SELECT queries simplified to only fetch existing columns
3. Model updated to remove non-existent fields
4. Using `current_step` instead of `current_day` everywhere
5. Removed all references to `last_message_at` and `last_sent_at`
6. UPDATE queries now only modify existing columns
