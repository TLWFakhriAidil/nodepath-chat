#!/bin/bash
# Fix WhatsApp Database Script

echo "Fixing WhatsApp Database Issues..."

# Get your database URL from environment or config
DATABASE_URL="${DATABASE_URL:-postgresql://user:pass@localhost/dbname}"

# Run the SQL fixes
psql "$DATABASE_URL" < fix_database_issues.sql

echo "Database fixes completed!"
echo ""
echo "Next steps:"
echo "1. Restart your WhatsApp application"
echo "2. The app should now start without crashing"
echo ""
echo "If you still have issues, check:"
echo "- The logs for any remaining errors"
echo "- Run: psql '$DATABASE_URL' -c 'SELECT column_name FROM information_schema.columns WHERE table_name = '\''whatsapp_chats'\'';'"
echo "- This will show you what columns exist in the whatsapp_chats table"