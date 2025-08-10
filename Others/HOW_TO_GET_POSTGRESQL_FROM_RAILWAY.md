# How to Get PostgreSQL Connection from Railway

## Steps:

1. **Login to Railway Dashboard**
   - Go to https://railway.app/dashboard
   - Select your project

2. **Find PostgreSQL Service**
   - Look for the PostgreSQL service in your project
   - Click on it to open the service details

3. **Get Connection String**
   - Click on the "Connect" tab
   - You'll see several connection options:
     - `DATABASE_URL` (internal - don't use this)
     - `DATABASE_PUBLIC_URL` (use this one!)
   
4. **Copy the Public URL**
   The URL will look like:
   ```
   postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@viaduct.proxy.rlwy.net:49914/railway
   ```

5. **Add to .env file**
   ```env
   DB_URI=postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@viaduct.proxy.rlwy.net:49914/railway
   ```

## Important Notes:

- Use the **PUBLIC** URL, not the internal URL
- The internal URL (railway.internal) only works within Railway's network
- Make sure to include `?sslmode=require` if not already in the URL
- Keep your credentials secure and never commit them to git

## Testing the Connection:

After adding to .env, run:
```bash
python database_operations.py
```

This will verify both PostgreSQL and MySQL connections are working.
