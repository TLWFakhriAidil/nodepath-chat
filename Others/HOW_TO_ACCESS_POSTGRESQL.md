# How to Access PostgreSQL and Run SQL Commands

## Option 1: Using psql Command Line
```bash
# Basic connection
psql -U username -d database_name

# With full connection string
psql "postgresql://username:password@localhost:5432/database_name"

# Once connected, you can run SQL commands directly
```

## Option 2: Using pgAdmin (GUI)
1. Download pgAdmin from https://www.pgadmin.org/
2. Install and open pgAdmin
3. Add your server connection:
   - Host: your_host (localhost or remote)
   - Port: 5432 (default)
   - Database: your_database_name
   - Username: your_username
   - Password: your_password
4. Right-click on your database → Query Tool
5. Paste and run SQL commands

## Option 3: Using DBeaver (Free Universal Database Tool)
1. Download from https://dbeaver.io/
2. Create new PostgreSQL connection
3. Enter your connection details
4. Open SQL Editor and run queries

## Option 4: Using TablePlus (Simple GUI)
1. Download from https://tableplus.com/
2. Create new connection
3. Open SQL editor with Cmd+K (or Ctrl+K)
4. Run your SQL

## Option 5: Online PostgreSQL Clients
If your database is accessible remotely, you can use:
- Adminer (single PHP file)
- phpPgAdmin (web-based)

## Quick Commands to Run in psql:
```sql
-- Once connected to psql, run these:

-- List all tables
\dt

-- Describe a table structure
\d whatsapp_chats
\d whatsapp_messages

-- Run SQL directly
SELECT * FROM whatsapp_chats LIMIT 5;
```

## Getting Your Connection String
Your connection string is usually in one of these formats:
- `postgresql://username:password@host:port/database`
- `postgres://username:password@host:port/database`

Check your application's environment variables or config file for the DATABASE_URL.