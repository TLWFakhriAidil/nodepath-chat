# Database Migration Guide

## Overview
This guide helps you apply the missing database migrations for the NodePath Chat application, specifically for the billing system and user profile features.

## Prerequisites
- Access to your MySQL database
- Database connection credentials
- Admin/root privileges on the database

## Missing Migrations

### 1. Billing System Tables (Migration 000011)
Creates the billing tables with `_nodepath` suffix:
- `subscriptions_nodepath` - User subscription plans
- `payments_nodepath` - Payment transactions with Billplz integration  
- `billing_history_nodepath` - Simplified billing records for display

### 2. User Profile Fields (Migration 000012)
Adds new columns to `user_nodepath` table:
- `gmail` - User's Gmail address (optional)
- `phone` - User's phone number (optional)  
- `status` - User account status (active/inactive)
- `expired` - Account expiration date (optional)

## How to Apply Migrations

### Option 1: Using the Migration Script (Recommended)
1. Navigate to the `scripts/` directory
2. Open `apply_migrations.sql` 
3. Copy the SQL commands
4. Execute them in your MySQL database:

```bash
# Connect to your database
mysql -h your_host -u your_username -p your_database_name

# Run the migration script
SOURCE /path/to/scripts/apply_migrations.sql;
```

### Option 2: Manual SQL Execution
Copy and paste the SQL commands from `scripts/apply_migrations.sql` directly into your MySQL client.

### Option 3: Using Migration Files
If you have a migration tool, apply these files in order:
1. `migrations/000011_create_billing_nodepath.up.sql`
2. `migrations/000012_add_user_profile_fields.up.sql`

## Verification

After applying the migrations, verify they worked correctly:

```sql
-- Check if billing tables were created
SHOW TABLES LIKE '%_nodepath';

-- Check user_nodepath table structure
DESCRIBE user_nodepath;

-- Check test subscription data
SELECT * FROM subscriptions_nodepath;

-- Verify user profile fields
SELECT id, email, full_name, gmail, phone, status, expired 
FROM user_nodepath 
LIMIT 5;
```

Expected results:
- You should see 3 new billing tables: `subscriptions_nodepath`, `payments_nodepath`, `billing_history_nodepath`
- `user_nodepath` should have new columns: `gmail`, `phone`, `status`, `expired`
- At least one test subscription record should exist

## Post-Migration Steps

1. **Restart your application** to ensure changes take effect
2. **Test the billing page** - navigate to `/billing` and verify it loads without errors
3. **Test the profile page** - navigate to `/profile` and verify you can view/edit your profile
4. **Check system status** - the top bar should show "System Online" based on your user status

## Troubleshooting

### Error: Table already exists
If you get "table already exists" errors, the tables may have been created with different names. Check:
```sql
SHOW TABLES LIKE '%subscription%';
SHOW TABLES LIKE '%payment%';
SHOW TABLES LIKE '%billing%';
```

### Error: Column already exists  
If you get "column already exists" errors for user profile fields, check:
```sql
DESCRIBE user_nodepath;
```

### Billing page still shows errors
1. Verify table names match exactly (with `_nodepath` suffix)
2. Check that test data was inserted
3. Restart your application server
4. Check application logs for specific error messages

### Profile page not accessible
1. Verify the `gmail`, `phone`, `status`, `expired` columns were added
2. Ensure existing users have `status = 'active'`
3. Restart your application server

## Database Schema Summary

### Billing Tables Structure
```sql
subscriptions_nodepath:
- id (VARCHAR 255, PRIMARY KEY)
- user_id (VARCHAR 255, NOT NULL)
- plan_name (VARCHAR 255, DEFAULT 'Test Plan')  
- plan_price (DECIMAL 10,2, DEFAULT 1.00)
- plan_period (VARCHAR 50, DEFAULT 'monthly')
- status (ENUM: active, cancelled, suspended, pending)
- next_billing_date (DATE, NOT NULL)
- features (JSON)
- created_at, updated_at (TIMESTAMP)

payments_nodepath:
- id (VARCHAR 255, PRIMARY KEY)
- user_id (VARCHAR 255, NOT NULL)
- subscription_id (VARCHAR 255, FK to subscriptions_nodepath)
- bill_id (VARCHAR 255, Billplz bill ID)
- invoice_number (VARCHAR 255, UNIQUE)
- amount (DECIMAL 10,2, NOT NULL)
- currency (VARCHAR 10, DEFAULT 'MYR')
- description (TEXT)
- status (ENUM: pending, paid, failed, cancelled)
- payment_method (VARCHAR 50, DEFAULT 'billplz')
- billplz_url (TEXT, Payment URL)
- paid_at (TIMESTAMP, NULL)
- created_at, updated_at (TIMESTAMP)

billing_history_nodepath:
- id (VARCHAR 255, PRIMARY KEY)
- user_id (VARCHAR 255, NOT NULL)
- payment_id (VARCHAR 255, FK to payments_nodepath)
- invoice_number (VARCHAR 255, NOT NULL)
- amount (DECIMAL 10,2, NOT NULL)
- currency (VARCHAR 10, DEFAULT 'MYR')
- description (TEXT, NOT NULL)
- status (ENUM: pending, paid, failed, cancelled)
- payment_date (DATE)
- created_at (TIMESTAMP)
```

### User Profile Fields Added
```sql
user_nodepath (additional columns):
- gmail (VARCHAR 255, NULL) - User's Gmail address
- phone (VARCHAR 20, NULL) - User's phone number
- status (VARCHAR 20, DEFAULT 'active') - Account status  
- expired (TIMESTAMP, NULL) - Account expiration date
```

## Support

If you encounter issues with the migration:
1. Check the application logs for specific error messages
2. Verify your database connection and permissions
3. Ensure you're using MySQL (not PostgreSQL) syntax
4. Make sure the database user has CREATE and ALTER privileges

The migration script uses `IF NOT EXISTS` and `IGNORE` clauses to prevent errors if tables/columns already exist, making it safe to run multiple times.