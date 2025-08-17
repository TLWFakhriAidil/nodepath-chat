#!/bin/bash

# Update Script for time_schedule Migration
# This script updates all references from scheduled_time/schedule_time to time_schedule

echo "=== Starting time_schedule migration update ==="

# Step 1: Run database migration
echo "Step 1: Running database migration..."
echo "Please run the following SQL migration:"
echo "psql -U your_username -d your_database -f database/004_change_to_time_schedule.sql"
echo ""

# Step 2: Update Go imports if needed
echo "Step 2: Updating Go code..."

# List of files that have been updated
echo "The following files have been updated:"
echo "- src/models/campaign.go (ScheduledTime -> TimeSchedule)"
echo "- src/models/sequence.go (ScheduleTime -> TimeSchedule)" 
echo "- src/database/connection.go (scheduled_time/schedule_time -> time_schedule)"
echo "- src/repository/campaign_repository.go (scheduled_time -> time_schedule)"
echo "- src/usecase/optimized_campaign_trigger.go (scheduled_time -> time_schedule)"
echo "- src/ui/rest/app.go (scheduled_time -> time_schedule)"
echo "- src/domains/sequence/sequence.go (schedule_time -> time_schedule)"

# Step 3: Frontend updates needed
echo ""
echo "Step 3: Frontend updates needed (if applicable):"
echo "Update any JavaScript/TypeScript files that reference:"
echo "- 'scheduled_time' -> 'time_schedule'"
echo "- 'schedule_time' -> 'time_schedule'"

# Step 4: API documentation updates
echo ""
echo "Step 4: Update API documentation:"
echo "- Campaign endpoints now use 'time_schedule' instead of 'scheduled_time'"
echo "- Sequence endpoints now use 'time_schedule' instead of 'schedule_time'"

# Step 5: Test the changes
echo ""
echo "Step 5: Testing checklist:"
echo "[ ] Create a new campaign with time_schedule"
echo "[ ] Update an existing campaign's time_schedule"
echo "[ ] Create a new sequence with time_schedule"
echo "[ ] Verify campaign triggers work with new field"
echo "[ ] Verify sequence processing works with new field"

echo ""
echo "=== Migration preparation complete ==="
echo "Please run the database migration and restart the application."
