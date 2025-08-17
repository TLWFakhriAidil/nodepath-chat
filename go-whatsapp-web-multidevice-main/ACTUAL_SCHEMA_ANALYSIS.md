# Database Schema Update for Anti-Spam Features

Based on your actual database schema, I can see the following tables and structure:

## Sequence-related tables:
1. **sequences** table with fields:
   - id
   - user_id
   - device_id
   - name
   - description
   - niche
   - status
   - created_at
   - updated_at
   - min_delay_seconds
   - max_delay_seconds
   - schedule_time
   - target_status
   - trigger

2. **sequence_contacts** table with fields:
   - id
   - sequence_id
   - contact_phone
   - contact_name
   - current_step
   - status
   - auto_enroll
   - skip_weekends
   - created_at
   - updated_at
   - total_days
   - active
   - schedule_time
   - min_delay_seconds
   - max_delay_seconds
   - total_contacts
   - active_contacts
   - completed_contacts
   - failed_contacts
   - progress_percentage
   - last_activity_at
   - estimated_completion_at
   - start_trigger
   - end_trigger
   - trigger

3. **sequence_steps** table with fields:
   - id
   - sequence_id
   - day_number
   - content
   - media_url
   - caption
   - delay_days
   - time_schedule
   - trigger
   - next_trigger
   - trigger_delay_hours
   - is_entry_point
   - min_delay_seconds
   - max_delay_seconds
   - send_time
   - is_ai

4. **user_devices** table with fields:
   - id
   - user_id
   - device_name
   - phone
   - jid
   - status
   - qr_code
   - created_at
   - updated_at
   - min_delay_seconds
   - max_delay_seconds
   - whacenter_instance
   - platform

The schema already has the necessary fields for anti-spam features!