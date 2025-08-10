# 🚀 Railway Auto-Deployment Guide

## Quick Deploy (One-Click)

### Option 1: Using Railway Button
[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/new/template?template=https%3A%2F%2Fgithub.com%2Faqilrvsb%2FWas-MCP&plugins=postgresql&envs=APP_PORT%2CAPP_DEBUG%2CAPP_OS%2CAPP_BASIC_AUTH%2CAPP_CHAT_FLUSH_INTERVAL%2CWHATSAPP_CHAT_STORAGE%2CWHATSAPP_ACCOUNT_VALIDATION%2CWHATSAPP_AUTO_REPLY&APP_PORTDesc=Application+port&APP_DEBUGDesc=Enable+debug+logging&APP_OSDesc=Device+name+shown+in+WhatsApp&APP_BASIC_AUTHDesc=Admin+credentials&APP_CHAT_FLUSH_INTERVALDesc=Days+to+keep+chats&WHATSAPP_CHAT_STORAGEDesc=Enable+chat+storage&WHATSAPP_ACCOUNT_VALIDATIONDesc=Validate+WhatsApp+accounts&WHATSAPP_AUTO_REPLYDesc=Auto+reply+message&APP_PORTDefault=3000&APP_DEBUGDefault=false&APP_OSDefault=WhatsApp+Business+System&APP_BASIC_AUTHDefault=admin%3Achangeme123&APP_CHAT_FLUSH_INTERVALDefault=30&WHATSAPP_CHAT_STORAGEDefault=true&WHATSAPP_ACCOUNT_VALIDATIONDefault=true&WHATSAPP_AUTO_REPLYDefault=Thank+you+for+contacting+us.+We+will+respond+shortly.)

### Option 2: Manual Setup with Scripts

#### For Windows:
```bash
# Run the auto-setup script
railway-auto-setup.bat
```

#### For Linux/Mac:
```bash
# Make executable and run
chmod +x railway-auto-setup.sh
./railway-auto-setup.sh
```

## 📋 Environment Variables (Auto-Configured)

The scripts will automatically set up these variables:

| Variable | Value | Description |
|----------|-------|-------------|
| `DB_URI` | Auto from DATABASE_URL | PostgreSQL connection |
| `APP_PORT` | 3000 | Application port |
| `APP_DEBUG` | false | Debug logging |
| `APP_OS` | WhatsApp Business System | Device name |
| `APP_BASIC_AUTH` | admin:changeme123 | Admin login |
| `APP_CHAT_FLUSH_INTERVAL` | 30 | Chat retention days |
| `WHATSAPP_CHAT_STORAGE` | true | Store chat history |
| `WHATSAPP_ACCOUNT_VALIDATION` | true | Validate accounts |
| `WHATSAPP_AUTO_REPLY` | Custom message | Auto-reply text |

## 🔧 Manual Railway Setup

If you prefer manual setup:

### 1. Create New Project
```bash
railway new
```

### 2. Add PostgreSQL
```bash
railway add postgresql
```

### 3. Deploy from GitHub
```bash
railway link
railway up
```

### 4. Set Environment Variables
```bash
# Database (automatic)
railway variables set DB_URI="$DATABASE_URL"

# Core settings
railway variables set APP_PORT=3000
railway variables set APP_DEBUG=false
railway variables set APP_OS="WhatsApp Business System"
railway variables set APP_BASIC_AUTH="admin:changeme123"

# WhatsApp features
railway variables set WHATSAPP_CHAT_STORAGE=true
railway variables set WHATSAPP_ACCOUNT_VALIDATION=true
railway variables set WHATSAPP_AUTO_REPLY="Thank you for contacting us."
railway variables set APP_CHAT_FLUSH_INTERVAL=30

# Optional webhook
railway variables set WHATSAPP_WEBHOOK="https://your-webhook.com"
railway variables set WHATSAPP_WEBHOOK_SECRET="your-secret"
```

## 🗄️ Database Auto-Setup

The application automatically creates these tables on startup:

- `users` - User accounts
- `user_devices` - WhatsApp devices
- `user_sessions` - Active sessions
- `campaigns` - Marketing campaigns
- `message_analytics` - Message tracking
- `whatsapp_chats` - Chat metadata
- `whatsapp_messages` - Message history
- `leads` - Lead management

## ✅ Post-Deployment Checklist

1. **Access your app**: 
   - URL: `https://your-app.up.railway.app`
   - Login: `admin@whatsapp.com` / `changeme123`

2. **Add WhatsApp devices**:
   - Go to Devices tab
   - Click "Add Device"
   - Scan QR code with WhatsApp

3. **Test features**:
   - Send a test message
   - Check analytics dashboard
   - View chat history

4. **Configure webhooks** (optional):
   - Add webhook URL in Railway variables
   - Test with webhook.site

## 🚨 Troubleshooting

### Build Fails
- Check logs: `railway logs`
- Ensure all imports are used
- Verify Go version compatibility

### Database Connection
- DATABASE_URL is automatically provided
- Check if PostgreSQL plugin is added
- Verify DB_URI references $DATABASE_URL

### WhatsApp Connection
- Ensure device has internet
- Try logout and rescan QR
- Check device status in dashboard

## 🎯 Optimization for 3000+ Devices

For your scale (200 users × 15 devices):

1. **Upgrade Railway Plan**:
   - Pro plan recommended
   - More CPU and memory

2. **Database Optimization**:
   ```sql
   -- Run these after deployment
   CREATE INDEX CONCURRENTLY idx_messages_device_timestamp 
   ON whatsapp_messages(device_id, timestamp);
   
   CREATE INDEX CONCURRENTLY idx_analytics_user_date 
   ON message_analytics(user_id, created_at);
   ```

3. **Environment Tweaks**:
   ```bash
   railway variables set APP_SHARD_COUNT=32
   railway variables set APP_CONNECTION_POOL=100
   railway variables set APP_MESSAGE_BUFFER=50
   ```

## 📞 Support

- **Issues**: Create GitHub issue
- **Railway**: Check Railway dashboard
- **Logs**: `railway logs --tail`

---

**Ready to deploy?** Run `railway-auto-setup.bat` (Windows) or `./railway-auto-setup.sh` (Linux/Mac)!
