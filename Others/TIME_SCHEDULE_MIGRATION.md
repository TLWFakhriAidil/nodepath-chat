# Time Schedule Migration Summary

## Overview
This migration changes the column name from `scheduled_time`/`schedule_time` to `time_schedule` for both campaigns and sequences tables to provide a unified naming convention.

## Database Changes

### 1. Campaigns Table
- Old column: `scheduled_time VARCHAR(10)`
- New column: `time_schedule TEXT`

### 2. Sequences Table  
- Old column: `schedule_time VARCHAR(10)`
- New column: `time_schedule TEXT`

### 3. Sequence Steps Table
- Old column: `schedule_time VARCHAR(10)`
- New column: `time_schedule TEXT`

## Files Updated

### Backend (Go)
1. **Models**
   - `src/models/campaign.go` - Changed field from `ScheduledTime` to `TimeSchedule`
   - `src/models/sequence.go` - Changed field from `ScheduleTime` to `TimeSchedule`

2. **Database Schema**
   - `src/database/connection.go` - Updated CREATE TABLE and ALTER TABLE statements

3. **Repository**
   - `src/repository/campaign_repository.go` - Updated all queries to use `time_schedule`

4. **Use Cases**
   - `src/usecase/optimized_campaign_trigger.go` - Updated campaign trigger logic

5. **REST API**
   - `src/ui/rest/app.go` - Updated request/response handling for campaigns

6. **Domain**
   - `src/domains/sequence/sequence.go` - Updated sequence domain structures

### Frontend
1. **Dashboard**
   - `src/views/dashboard.html` - Updated JavaScript to use `time_schedule` (8 occurrences)

## Migration Steps

1. **Stop the application**

2. **Run the database migration**
   ```bash
   psql -U your_username -d your_database -f database/004_change_to_time_schedule.sql
   ```

3. **Verify migration success**
   ```sql
   -- Check campaigns table
   \d campaigns
   
   -- Check sequences table  
   \d sequences
   
   -- Check sequence_steps table
   \d sequence_steps
   ```

4. **Restart the application**

5. **Test functionality**
   - Create a new campaign with time schedule
   - Create a new sequence with time schedule
   - Verify existing campaigns still work
   - Verify campaign triggers execute properly
   - Verify sequence processing works

## API Changes

### Campaign Endpoints
**Before:**
```json
{
  "scheduled_time": "14:30"
}
```

**After:**
```json
{
  "time_schedule": "14:30"
}
```

### Sequence Endpoints
**Before:**
```json
{
  "schedule_time": "09:00"
}
```

**After:**
```json
{
  "time_schedule": "09:00"
}
```

## Rollback Plan

If you need to rollback:

1. **Database rollback**
   ```sql
   -- Campaigns
   ALTER TABLE campaigns ADD COLUMN scheduled_time VARCHAR(10);
   UPDATE campaigns SET scheduled_time = time_schedule;
   ALTER TABLE campaigns DROP COLUMN time_schedule;
   
   -- Sequences
   ALTER TABLE sequences ADD COLUMN schedule_time VARCHAR(10);
   UPDATE sequences SET schedule_time = time_schedule;
   ALTER TABLE sequences DROP COLUMN time_schedule;
   
   -- Sequence steps
   ALTER TABLE sequence_steps ADD COLUMN schedule_time VARCHAR(10);
   UPDATE sequence_steps SET schedule_time = time_schedule;
   ALTER TABLE sequence_steps DROP COLUMN time_schedule;
   ```

2. **Code rollback**
   - Revert all Go file changes
   - Revert dashboard.html changes

## Benefits of This Change

1. **Consistency** - Single field name across all tables
2. **Flexibility** - TEXT type allows for more complex scheduling in future
3. **Clarity** - `time_schedule` is more descriptive than `scheduled_time`
4. **Maintainability** - Easier to search and maintain with consistent naming

## Notes

- The migration preserves all existing data
- NULL/empty values continue to mean "run immediately"
- Time format validation remains the same (HH:MM or HH:MM:SS)
- Timezone handling through `scheduled_at` column remains unchanged
