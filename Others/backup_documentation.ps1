# PowerShell script to backup Railway PostgreSQL database
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "Railway Database Backup - Working Version" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

$backupDir = "backups\2025-07-01_00-01-03_working_version"
$dbUrl = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

Write-Host "Creating backup documentation..." -ForegroundColor Yellow

# Create backup info file
$backupInfo = @{
    "backup_date" = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    "database_url" = $dbUrl
    "git_commit" = git log --oneline -1
    "working_features" = @(
        "Device Management - 6 devices per row layout",
        "Campaign Clone and Delete icons at top",
        "Lead Management accessible without WhatsApp connection",
        "Campaign calendar with status colors",
        "Broadcast message processing with human-like delays"
    )
}

$backupInfo | ConvertTo-Json -Depth 3 | Out-File "$backupDir\backup_info.json" -Encoding UTF8

# Create restore instructions
@"
RESTORE INSTRUCTIONS
===================

This backup was created on: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")
From commit: $(git log --oneline -1)

To restore this database:

1. Install PostgreSQL client tools if not already installed:
   - Download from: https://www.postgresql.org/download/windows/
   - Or use: choco install postgresql

2. Use one of these methods:

   Method A - Using psql:
   psql "$dbUrl" < postgresql_backup.sql

   Method B - Using Railway CLI:
   railway run psql < postgresql_backup.sql

   Method C - Manual restore via Railway dashboard:
   - Go to your Postgres service
   - Use the query interface to run the SQL

3. After restore, restart your application:
   railway restart

IMPORTANT NOTES:
- This will REPLACE all current data
- Make sure to backup current data first
- The application should be stopped during restore

Database Statistics at backup time:
- Check database_stats.json for table counts
- Check backup_info.json for system state
"@ | Out-File "$backupDir\RESTORE_INSTRUCTIONS.txt" -Encoding UTF8

Write-Host "Backup documentation created!" -ForegroundColor Green
Write-Host ""
Write-Host "Files created in $backupDir`:" -ForegroundColor White
Write-Host "- backup_info.json (system state and configuration)" -ForegroundColor Gray
Write-Host "- RESTORE_INSTRUCTIONS.txt (how to restore)" -ForegroundColor Gray
Write-Host ""

# Try to get table counts using available tools
Write-Host "Attempting to get database statistics..." -ForegroundColor Yellow

# Create a simple stats file
@"
DATABASE STATISTICS
==================
Backup Date: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

Git Commit: $(git log --oneline -1)

Key Tables:
- users (user accounts and sessions)
- devices (WhatsApp device connections)
- leads (contact information)
- campaigns (broadcast campaigns)
- broadcast_messages (message queue)
- sequences (automated message sequences)
- whatsapp_chats (chat history)
- whatsapp_messages (individual messages)

Connection Info:
Host: yamanote.proxy.rlwy.net
Port: 49914
Database: railway
SSL: required

To get current counts, run these queries in Railway:
SELECT 'users', COUNT(*) FROM users;
SELECT 'devices', COUNT(*) FROM devices;
SELECT 'leads', COUNT(*) FROM leads;
SELECT 'campaigns', COUNT(*) FROM campaigns;
SELECT 'broadcast_messages', COUNT(*) FROM broadcast_messages;
"@ | Out-File "$backupDir\database_info.txt" -Encoding UTF8

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "Backup documentation completed!" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "IMPORTANT: To create actual SQL backup, you need to:" -ForegroundColor Yellow
Write-Host "1. Install PostgreSQL tools: https://www.postgresql.org/download/" -ForegroundColor White
Write-Host "2. Run: pg_dump `"$dbUrl`" > `"$backupDir\postgresql_backup.sql`"" -ForegroundColor White
Write-Host ""
Write-Host "Or use Railway's dashboard to export data manually." -ForegroundColor White
Write-Host ""
Write-Host "All backup documentation saved in: $backupDir" -ForegroundColor Green
