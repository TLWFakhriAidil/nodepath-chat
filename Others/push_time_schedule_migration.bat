@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo === Running Time Schedule Migration ===
echo.

echo Step 1: Adding files to git...
git add database/004_change_to_time_schedule.sql
git add src/models/campaign.go
git add src/models/sequence.go
git add src/database/connection.go
git add src/repository/campaign_repository.go
git add src/usecase/optimized_campaign_trigger.go
git add src/ui/rest/app.go
git add src/domains/sequence/sequence.go
git add src/views/dashboard.html
git add TIME_SCHEDULE_MIGRATION.md
git add update_time_schedule.sh

echo.
echo Step 2: Committing changes...
git commit -m "feat: Change scheduled_time/schedule_time to time_schedule

- Add migration 004_change_to_time_schedule.sql
- Update all Go models and repositories
- Update campaign and sequence logic
- Update frontend dashboard
- Unified naming convention across system
- Preserve all existing data with safe migration"

echo.
echo Step 3: Pulling latest changes...
git pull origin main --rebase

echo.
echo Step 4: Pushing to main branch...
git push origin main

echo.
echo === Migration files pushed successfully ===
echo.
echo Now running database migrations...
echo Please ensure your DB_URI is set correctly.
pause
