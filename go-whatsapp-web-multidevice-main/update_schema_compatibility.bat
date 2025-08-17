@echo off
echo ========================================
echo Updating Code to Match Actual Database Schema
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

echo Creating schema compatibility fixes...

REM Create a SQL file to add any missing columns as aliases if needed
echo -- Schema Compatibility Layer > schema_compatibility.sql
echo -- This ensures the code works with the actual database schema >> schema_compatibility.sql
echo. >> schema_compatibility.sql
echo -- Check if next_trigger_time exists, if not create it as an alias or column >> schema_compatibility.sql
echo DO $$ >> schema_compatibility.sql
echo BEGIN >> schema_compatibility.sql
echo     -- Add next_trigger_time if it doesn't exist (for sequence processing) >> schema_compatibility.sql
echo     IF NOT EXISTS (SELECT 1 FROM information_schema.columns >> schema_compatibility.sql
echo                    WHERE table_name = 'sequence_contacts' >> schema_compatibility.sql
echo                    AND column_name = 'next_trigger_time') THEN >> schema_compatibility.sql
echo         ALTER TABLE sequence_contacts ADD COLUMN next_trigger_time TIMESTAMP; >> schema_compatibility.sql
echo     END IF; >> schema_compatibility.sql
echo. >> schema_compatibility.sql
echo     -- Add current_trigger if it doesn't exist >> schema_compatibility.sql
echo     IF NOT EXISTS (SELECT 1 FROM information_schema.columns >> schema_compatibility.sql
echo                    WHERE table_name = 'sequence_contacts' >> schema_compatibility.sql
echo                    AND column_name = 'current_trigger') THEN >> schema_compatibility.sql
echo         ALTER TABLE sequence_contacts ADD COLUMN current_trigger VARCHAR(255); >> schema_compatibility.sql
echo     END IF; >> schema_compatibility.sql
echo. >> schema_compatibility.sql
echo     -- Add processing_device_id if it doesn't exist >> schema_compatibility.sql
echo     IF NOT EXISTS (SELECT 1 FROM information_schema.columns >> schema_compatibility.sql
echo                    WHERE table_name = 'sequence_contacts' >> schema_compatibility.sql
echo                    AND column_name = 'processing_device_id') THEN >> schema_compatibility.sql
echo         ALTER TABLE sequence_contacts ADD COLUMN processing_device_id UUID; >> schema_compatibility.sql
echo     END IF; >> schema_compatibility.sql
echo. >> schema_compatibility.sql
echo     -- Add sequence_stepid if it doesn't exist >> schema_compatibility.sql
echo     IF NOT EXISTS (SELECT 1 FROM information_schema.columns >> schema_compatibility.sql
echo                    WHERE table_name = 'sequence_contacts' >> schema_compatibility.sql
echo                    AND column_name = 'sequence_stepid') THEN >> schema_compatibility.sql
echo         ALTER TABLE sequence_contacts ADD COLUMN sequence_stepid UUID; >> schema_compatibility.sql
echo     END IF; >> schema_compatibility.sql
echo. >> schema_compatibility.sql
echo     -- Add completed_at if it doesn't exist >> schema_compatibility.sql
echo     IF NOT EXISTS (SELECT 1 FROM information_schema.columns >> schema_compatibility.sql
echo                    WHERE table_name = 'sequence_contacts' >> schema_compatibility.sql
echo                    AND column_name = 'completed_at') THEN >> schema_compatibility.sql
echo         ALTER TABLE sequence_contacts ADD COLUMN completed_at TIMESTAMP; >> schema_compatibility.sql
echo     END IF; >> schema_compatibility.sql
echo. >> schema_compatibility.sql
echo     -- Add processing_started_at if it doesn't exist >> schema_compatibility.sql
echo     IF NOT EXISTS (SELECT 1 FROM information_schema.columns >> schema_compatibility.sql
echo                    WHERE table_name = 'sequence_contacts' >> schema_compatibility.sql
echo                    AND column_name = 'processing_started_at') THEN >> schema_compatibility.sql
echo         ALTER TABLE sequence_contacts ADD COLUMN processing_started_at TIMESTAMP; >> schema_compatibility.sql
echo     END IF; >> schema_compatibility.sql
echo. >> schema_compatibility.sql
echo     -- Add last_error if it doesn't exist >> schema_compatibility.sql
echo     IF NOT EXISTS (SELECT 1 FROM information_schema.columns >> schema_compatibility.sql
echo                    WHERE table_name = 'sequence_contacts' >> schema_compatibility.sql
echo                    AND column_name = 'last_error') THEN >> schema_compatibility.sql
echo         ALTER TABLE sequence_contacts ADD COLUMN last_error TEXT; >> schema_compatibility.sql
echo     END IF; >> schema_compatibility.sql
echo. >> schema_compatibility.sql
echo     -- Add retry_count if it doesn't exist >> schema_compatibility.sql
echo     IF NOT EXISTS (SELECT 1 FROM information_schema.columns >> schema_compatibility.sql
echo                    WHERE table_name = 'sequence_contacts' >> schema_compatibility.sql
echo                    AND column_name = 'retry_count') THEN >> schema_compatibility.sql
echo         ALTER TABLE sequence_contacts ADD COLUMN retry_count INTEGER DEFAULT 0; >> schema_compatibility.sql
echo     END IF; >> schema_compatibility.sql
echo. >> schema_compatibility.sql
echo     -- Add assigned_device_id if it doesn't exist >> schema_compatibility.sql
echo     IF NOT EXISTS (SELECT 1 FROM information_schema.columns >> schema_compatibility.sql
echo                    WHERE table_name = 'sequence_contacts' >> schema_compatibility.sql
echo                    AND column_name = 'assigned_device_id') THEN >> schema_compatibility.sql
echo         ALTER TABLE sequence_contacts ADD COLUMN assigned_device_id UUID; >> schema_compatibility.sql
echo     END IF; >> schema_compatibility.sql
echo END$$; >> schema_compatibility.sql

echo.
echo Schema compatibility SQL file created.
echo.
echo To apply these changes to your database:
echo 1. Connect to your PostgreSQL database
echo 2. Run: \i schema_compatibility.sql
echo.
echo Or run this command:
echo psql -U your_user -d your_database -f schema_compatibility.sql
echo.
echo The anti-spam features are already integrated into the code!
echo.
echo Key points:
echo - Your schema already has min_delay_seconds and max_delay_seconds in sequences table
echo - The code will use these delays for human-like behavior
echo - Malaysian greetings will be applied automatically
echo - Message randomization is built into the WhatsApp sender
echo.
pause
