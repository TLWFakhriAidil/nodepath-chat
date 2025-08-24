# Database Connection Issue Fix

## Problem Identified

The application is experiencing "database not available" errors because the current IP address `113.211.138.14` is not whitelisted on the MySQL server at `159.89.198.71:3306`.

### Error Details
- **Error**: `Error 1130: Host '113.211.138.14' is not allowed to connect to this MySQL server`
- **Database Server**: `159.89.198.71:3306`
- **Database**: `admin_railway`
- **Current IP**: `113.211.138.14`
- **Impact**: All database operations fail, causing "database not available" errors in the application

## Solutions

### Solution 1: Whitelist Current IP Address (Recommended)

**For Database Administrator:**
1. Connect to the MySQL server at `159.89.198.71:3306`
2. Run the following SQL commands:

```sql
-- Grant access to the current IP address
GRANT ALL PRIVILEGES ON admin_railway.* TO 'admin_aqil'@'113.211.138.14' IDENTIFIED BY 'admin_aqil';

-- Flush privileges to apply changes
FLUSH PRIVILEGES;

-- Verify the user was created
SELECT User, Host FROM mysql.user WHERE User = 'admin_aqil';
```

### Solution 2: Allow Access from Any IP (Less Secure)

**Warning**: This is less secure but works for development/testing.

```sql
-- Grant access from any IP address
GRANT ALL PRIVILEGES ON admin_railway.* TO 'admin_aqil'@'%' IDENTIFIED BY 'admin_aqil';

-- Flush privileges to apply changes
FLUSH PRIVILEGES;
```

### Solution 3: Use IP Range (Balanced Security)

If you know the IP range your application uses:

```sql
-- Grant access to IP range (example: 113.211.138.0/24)
GRANT ALL PRIVILEGES ON admin_railway.* TO 'admin_aqil'@'113.211.138.%' IDENTIFIED BY 'admin_aqil';

-- Flush privileges to apply changes
FLUSH PRIVILEGES;
```

## Testing the Fix

After applying any of the above solutions, test the connection:

```bash
# Run the database connection test
go run test_db_connection.go
```

Expected output after fix:
```
✅ Database connection successful!
✅ Database version: 5.7.x
✅ chatbot_flows_nodepath table exists with X rows
🎉 All database tests completed successfully!
```

## Restart Application

After fixing the database connection, restart the application:

```bash
# Stop the current server (Ctrl+C)
# Then restart
go run cmd/server/main.go
```

The application should now show:
```
time="..." level=info msg="Database connection established successfully"
time="..." level=info msg="Database migrations completed"
```

## Verification

1. **Check server logs**: No more "database not available" errors
2. **Test API endpoints**: `/api/flows` should return data instead of 500 errors
3. **Check health endpoint**: `/api/health` should show database as healthy

## Security Recommendations

1. **Use specific IP addresses** when possible instead of wildcards
2. **Regularly review** MySQL user permissions
3. **Consider using SSL** for database connections in production
4. **Monitor connection logs** for unauthorized access attempts

## Alternative: Local Database Setup

If you cannot modify the remote MySQL server, you can set up a local database:

1. Install MySQL locally
2. Create the database and tables
3. Update `.env` file with local connection details
4. Run migrations to create tables

```env
# Local database configuration
MYSQL_URI=mysql://root:password@localhost:3306/nodepath_chat
```

## Contact Information

If you need help with database administration, contact the database administrator with:
- **Current IP**: `113.211.138.14`
- **Required Access**: Full privileges on `admin_railway` database
- **User**: `admin_aqil`
- **Purpose**: NodePath Chat application database access