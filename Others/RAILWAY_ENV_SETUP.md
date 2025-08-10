# Railway Environment Variables Configuration

## Required Environment Variables:

1. **DB_URI** (IMPORTANT)
   ```
   postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require
   ```
   (Use your DATABASE_PUBLIC_URL from PostgreSQL service)

2. **APP_PORT**
   ```
   3000
   ```

3. **APP_DEBUG**
   ```
   false
   ```

4. **APP_OS**
   ```
   Chrome
   ```

5. **APP_BASIC_AUTH**
   ```
   admin:your-secure-password
   ```
   (Change this to your desired username:password)

6. **APP_CHAT_FLUSH_INTERVAL**
   ```
   7
   ```

7. **WHATSAPP_AUTO_REPLY**
   ```
   Auto reply from Railway WhatsApp Bot
   ```
   (Optional - leave empty if you don't want auto-reply)

8. **WHATSAPP_WEBHOOK**
   ```
   https://your-webhook-url.com
   ```
   (Optional - leave empty if you don't use webhooks)

9. **WHATSAPP_WEBHOOK_SECRET**
   ```
   your-webhook-secret
   ```
   (Optional - but required if using webhooks)

10. **WHATSAPP_ACCOUNT_VALIDATION**
    ```
    true
    ```

11. **WHATSAPP_CHAT_STORAGE**
    ```
    true
    ```

## How to add in Railway:

1. Go to your Railway project
2. Select your service
3. Go to "Variables" tab
4. Add each variable one by one
5. Or use "Raw Editor" and paste all at once:

```
DB_URI=postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require
APP_PORT=3000
APP_DEBUG=false
APP_OS=Chrome
APP_BASIC_AUTH=admin:changeme123
APP_CHAT_FLUSH_INTERVAL=7
WHATSAPP_AUTO_REPLY=
WHATSAPP_WEBHOOK=
WHATSAPP_WEBHOOK_SECRET=
WHATSAPP_ACCOUNT_VALIDATION=true
WHATSAPP_CHAT_STORAGE=true
```