# Instructions to Add PostgreSQL Connection

Since you don't have a PostgreSQL URL configured, you need to:

1. **Get your PostgreSQL URL from Railway:**
   - Log in to https://railway.app/dashboard
   - Select your WhatsApp project
   - Click on the PostgreSQL service
   - Go to the "Connect" tab
   - Copy the `DATABASE_PUBLIC_URL` (NOT the internal URL)

2. **Add to your .env file:**
   Open your .env file and add:
   ```
   DB_URI=postgresql://postgres:YOUR_PASSWORD@YOUR_HOST.railway.app:PORT/railway
   ```

3. **Example PostgreSQL URLs:**
   ```
   # Railway format:
   DB_URI=postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@viaduct.proxy.rlwy.net:49914/railway
   
   # Local PostgreSQL:
   DB_URI=postgresql://postgres:password@localhost:5432/whatsapp_db
   
   # External PostgreSQL:
   DB_URI=postgresql://user:password@host.com:5432/database?sslmode=require
   ```

4. **Test the connection:**
   After adding the URL, run:
   ```
   python db_operations_fixed.py
   ```

## Important Notes:
- Use the PUBLIC URL from Railway, not the internal URL
- Include `?sslmode=require` for secure connections
- The PostgreSQL database stores WhatsApp session data
- MySQL stores all application data

## Need Help?
If you don't have a PostgreSQL database yet:
1. You can use Railway's free PostgreSQL service
2. Or install PostgreSQL locally
3. Or use any cloud PostgreSQL provider (AWS RDS, Google Cloud SQL, etc.)
