# 🗄️ WhatsApp Database Backup & Restore Guide

## 📋 Overview

This backup system helps you protect your WhatsApp Multi-Device database from data loss. It includes tools for:
- Creating backups (schema only, data only, or full)
- Restoring from backups
- Fixing missing columns
- Working with Railway PostgreSQL

## 🔧 Prerequisites

1. **PostgreSQL client tools** (pg_dump, psql)
   - Windows: Download from https://www.postgresql.org/download/windows/
   - Add to PATH: `C:\Program Files\PostgreSQL\XX\bin`

2. **Database credentials**
   - For local: host, port, database name, username, password
   - For Railway: DATABASE_URL from Railway dashboard

## 📦 Backup Files Created

### 1. **backup_whatsapp_schema.sql**
- Complete WhatsApp table schema
- All required columns (lid, facebook_uuid, etc.)
- Use this as reference for table structure

### 2. **Backup Scripts**

#### **backup_database.bat** - Local Database Backup
```cmd
backup_database.bat
```
- Choose backup type:
  1. Schema only (structure)
  2. Data only
  3. Full backup (structure + data)
- Creates timestamped backup files

#### **backup_from_railway.bat** - Railway Database Backup
```cmd
backup_from_railway.bat
```
- Backs up from Railway PostgreSQL
- Requires DATABASE_URL from Railway
- Creates local backup file

### 3. **Restore Scripts**

#### **restore_database.bat** - Local Database Restore
```cmd
restore_database.bat
```
Options:
1. Restore from backup file
2. Restore WhatsApp tables only
3. Quick fix - add missing columns

#### **restore_to_railway.bat** - Railway Database Restore
```cmd
restore_to_railway.bat
```
- Restores backup to Railway
- Option to drop tables first (clean restore)

## 🚀 Quick Start Guide

### Creating Your First Backup (Railway)

1. **Get your DATABASE_URL**:
   - Go to Railway dashboard
   - Click on Postgres service
   - Go to "Connect" tab
   - Copy DATABASE_URL

2. **Run backup script**:
   ```cmd
   backup_from_railway.bat
   ```

3. **Paste DATABASE_URL** when prompted

4. **Choose backup type** (1 for WhatsApp tables only)

5. **Save the backup file** safely!

### Restoring After Table Drop

If tables get dropped, restore them:

```cmd
restore_to_railway.bat
```
1. Select your backup file
2. Paste DATABASE_URL
3. Choose option 1 (clean restore)

## 📊 WhatsApp Tables Included

1. `whatsmeow_device` - Device information
2. `whatsmeow_identity_keys` - Identity keys
3. `whatsmeow_pre_keys` - Pre-keys
4. `whatsmeow_sessions` - Active sessions
5. `whatsmeow_sender_keys` - Sender keys
6. `whatsmeow_app_state_sync_keys` - Sync keys
7. `whatsmeow_app_state_version` - App state versions
8. `whatsmeow_app_state_mutation_macs` - Mutation MACs
9. `whatsmeow_contacts` - Contacts
10. `whatsmeow_chat_settings` - Chat settings
11. `whatsmeow_message_secrets` - Message secrets
12. `whatsmeow_privacy_tokens` - Privacy tokens

## 🛡️ Best Practices

1. **Regular Backups**
   - Daily for active systems
   - Before major updates
   - After adding new devices

2. **Backup Storage**
   - Keep multiple versions
   - Store offsite (cloud storage)
   - Test restore process regularly

3. **Quick Fixes**
   - Use option 3 in restore_database.bat for missing columns
   - Faster than full restore for minor issues

## 🚨 Troubleshooting

### "pg_dump not found"
- Install PostgreSQL client tools
- Add to system PATH

### "Connection refused"
- Check DATABASE_URL is correct
- Ensure you can access Railway database

### "Permission denied"
- Check database user has necessary privileges
- For Railway, use provided credentials

## 📝 Example Commands

### Manual backup (Railway):
```bash
pg_dump "postgresql://user:pass@host:port/db" -t "whatsmeow_*" > backup.sql
```

### Manual restore (Railway):
```bash
psql "postgresql://user:pass@host:port/db" < backup.sql
```

### Check if tables exist:
```sql
SELECT table_name 
FROM information_schema.tables 
WHERE table_name LIKE 'whatsmeow_%';
```

## 🔄 Automated Backup (Optional)

For automated daily backups, create a scheduled task:
1. Use Windows Task Scheduler
2. Run backup_from_railway.bat daily
3. Keep last 7 days of backups

---

**Remember**: Always test your restore process before you need it in an emergency!
