# Redis Environment Variables for Railway

Add these variables to your Railway project:

## Required Variables:

1. **REDIS_URL**
   ```
   ${{REDIS_URL}}
   ```

2. **REDIS_PASSWORD**
   ```
   ${{REDIS_PASSWORD}}
   ```

3. **REDIS_HOST**
   ```
   ${{REDIS_PRIVATE_DOMAIN}}
   ```

4. **REDIS_PORT**
   ```
   6379
   ```

## How to Add:

1. Click "New Variable" button
2. Add each variable name and value
3. Save changes

## Alternative Method:

If the Redis plugin variables are not showing, try:

1. Go to your Redis service (click the Redis icon)
2. Go to the "Connect" tab
3. You should see the connection details there
4. Copy the Redis URL and add it manually

## Manual Redis URL Format:

If you need to construct it manually:
```
redis://default:YOUR_REDIS_PASSWORD@YOUR_REDIS_HOST.railway.internal:6379
```

## Verify Redis Connection:

After adding the variables, your app logs should show:
```
Successfully connected to Redis
Redis URL found, initializing Redis-based broadcast manager
```

## Troubleshooting:

If variables still don't appear:
1. Remove the Redis service
2. Re-add it: Click "New" → "Database" → "Add Redis"
3. Railway should create the variables automatically

Or you can use the Railway CLI:
```bash
railway variables set REDIS_URL=${{REDIS_URL}}
railway variables set REDIS_PASSWORD=${{REDIS_PASSWORD}}
```
