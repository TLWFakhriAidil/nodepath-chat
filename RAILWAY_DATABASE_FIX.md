# Railway Database Connection Fix

## Issue Identified

The Railway deployment is experiencing "database not available" errors because the `DATABASE_URL` environment variable is configured to connect to a non-existent SSH tunnel service:

```
DATABASE_URL=mysql://admin_aqil:admin_aqil@mysql-ssh-tunnel.railway.internal:3307/admin_railway
```

The SSH tunnel service (`mysql-ssh-tunnel.railway.internal`) is not deployed or running, causing all database connections to fail.

## Root Cause

1. **Missing SSH Tunnel Service**: The Railway deployment expects an SSH tunnel service to be running, but it's not configured or deployed.
2. **Incorrect Environment Variables**: The `DATABASE_URL` points to a tunnel service that doesn't exist.
3. **Application Graceful Handling**: The Go application correctly handles database connection failures by setting `db = nil` and continuing to run, but all database-dependent endpoints return "database not available" errors.

## Solution

### Option 1: Direct Database Connection (Recommended)

Update the Railway environment variables to use direct database connection:

1. **Login to Railway Dashboard**
2. **Navigate to your project**
3. **Go to Variables tab**
4. **Update the following environment variables**:

```bash
DATABASE_URL=mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway?charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci
MYSQL_URI=mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway?charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci
```

5. **Redeploy the application**

### Option 2: Deploy SSH Tunnel Service (Advanced)

If you prefer to use SSH tunnel for security:

1. **Deploy SSH tunnel service** using the provided `Dockerfile.tunnel`
2. **Configure SSH credentials** and jump server
3. **Update environment variables** to use tunnel service

## Verification Steps

1. **Check Application Logs**:
   ```bash
   # Look for successful database connection
   "Database initialized successfully"
   "Database migrations completed"
   ```

2. **Test API Endpoints**:
   ```bash
   curl https://nodepath-chat-production.up.railway.app/api/flows
   ```

3. **Monitor Error Logs**:
   - No more "database not available" errors
   - No more "Failed to get flows" errors

## Files Updated

- `railway-deploy.yml`: Updated to use direct database connection
- `RAILWAY_DATABASE_FIX.md`: This documentation file

## Next Steps

1. **Update Railway environment variables** as described above
2. **Redeploy the application**
3. **Test the `/api/flows` endpoint**
4. **Monitor application logs** for successful database connection

## Notes

- The direct connection approach works because Railway allows outbound connections to external databases
- The MySQL server at `159.89.198.71:3306` must allow connections from Railway's IP ranges
- If IP whitelisting is required, consider using Railway's static IP feature or implement the SSH tunnel solution

## Contact

If you encounter issues with IP whitelisting or need assistance with SSH tunnel setup, please contact the database administrator or consider using Railway's managed database services.