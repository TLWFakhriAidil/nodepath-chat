# Schema Migration Control

## Current Status: MIGRATIONS DISABLED

The schema migrations have been disabled as requested. The system will connect to the database but will NOT run any schema creation or migration queries.

## How to Control Migrations

### To DISABLE migrations (current state):
The migrations are currently disabled in `src/database/connection.go`. The InitializeSchema() call is commented out.

### To ENABLE migrations:
1. Open `src/database/connection.go`
2. Find the section around line 41-47
3. Uncomment the InitializeSchema() call:
   ```go
   // Change from:
   /*
   if err := InitializeSchema(); err != nil {
       log.Fatalf("Failed to initialize schema: %v", err)
   }
   */
   
   // To:
   if err := InitializeSchema(); err != nil {
       log.Fatalf("Failed to initialize schema: %v", err)
   }
   ```

## Important Notes

- With migrations disabled, the system assumes your database schema is already correctly set up
- No tables will be created automatically
- No schema updates will be applied
- You must manage your database schema manually

## Required Tables

If you need to create the schema manually, the required tables are:
- users
- user_devices
- user_sessions
- campaigns
- sequences
- sequence_steps
- sequence_contacts
- broadcast_messages
- leads
- And other supporting tables

Refer to CURRENT_WORKING_SCHEMA.sql for the complete schema structure.
