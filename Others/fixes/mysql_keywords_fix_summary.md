# MySQL Reserved Keywords Fix - Summary

## Issues Found and Fixed

### 1. Lead Section Error (FIXED)
**Error:**
```
2025/07/30 06:21:34 Error getting leads: Error 1064 (42000): You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use near 'trigger, created_at, updated_at FROM leads WHERE user_id = ? AND device_id ='
```

**Cause:** The column name `trigger` is a reserved keyword in MySQL and needs to be escaped with backticks.

**Solution:** All references to the `trigger` column in `lead_repository.go` have been properly escaped with backticks (`` `trigger` ``).

**Status:** ✅ FIXED - All 8 references to trigger column are now properly escaped.

### 2. Queued Message Cleaner Error (FIXED)
**Error:**
```
time="2025-07-30T06:21:52Z" level=error msg="Failed to clean stuck messages: Error 1064 (42000): You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use near ')' at line 6"
```

**Cause:** Extra parenthesis in the SQL query in `queued_message_cleaner.go`.

**Original:**
```sql
AND updated_at < (DATE_SUB(CURRENT_TIMESTAMP, INTERVAL 12 HOUR))
```

**Fixed to:**
```sql
AND updated_at < DATE_SUB(CURRENT_TIMESTAMP, INTERVAL 12 HOUR)
```

**Status:** ✅ FIXED - The extra parenthesis has been removed.

## Database Architecture Understanding

### Dual Database System
1. **PostgreSQL** - Used for WhatsApp session storage (required by whatsmeow library)
   - Connection: `DB_URI=postgresql://...`
   - Tables: All `whatsmeow_*` tables

2. **MySQL** - Used for application data
   - Connection: `MYSQL_URI=mysql://...`
   - Tables: 33 tables including users, leads, campaigns, sequences, etc.

### Key MySQL Compatibility Changes
- Parameter placeholders: `$1, $2` → `?, ?`
- UUID generation: `gen_random_uuid()` → `UUID()`
- Boolean values: `TRUE/FALSE` → `1/0`
- Case-insensitive search: `ILIKE` → `LIKE`
- Reserved keywords: Must be escaped with backticks

## Next Steps

1. **Rebuild the application:**
   ```bash
   go build -o whatsapp.exe
   ```

2. **Restart the application:**
   ```bash
   whatsapp.exe rest
   ```

3. **Test the lead section:**
   - Navigate to the leads page
   - Try creating, viewing, and updating leads
   - Verify no SQL errors appear in the console

## Additional Notes

- The system uses a sophisticated dual-database architecture to separate WhatsApp sessions from application data
- All MySQL reserved keywords should be escaped with backticks when used as column names
- The fixes have been applied without changing the database schema
- Both errors were SQL syntax issues specific to MySQL compatibility

## Files Modified

1. `src/repository/lead_repository.go` - Ensured all `trigger` column references are escaped
2. `src/usecase/queued_message_cleaner.go` - Fixed SQL syntax error (removed extra parenthesis)

The application should now work correctly with the lead section without any SQL syntax errors.
