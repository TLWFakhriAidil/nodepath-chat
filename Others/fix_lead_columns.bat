@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo === Fixing lead repository database column mismatch ===
echo.

git add src/repository/lead_repository.go
git add src/database/connection.go
git add database/005_fix_leads_table.sql
git commit -m "fix: Fix lead repository to match actual database columns

- Fixed GetLeadsByNiche to use journey instead of email/source/notes
- Fixed GetNewLeadsForSequence to match query columns
- Added proper handling for journey to Notes field mapping
- Added auto-migration for target_status column
- Added logging for scan errors to help debug"

git push origin main

echo.
echo === Fix pushed! ===
echo Your campaign should now execute without column errors.
pause
