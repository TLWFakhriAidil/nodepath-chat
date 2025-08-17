@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo === Pushing Time Schedule Migration with Auto-Migration ===
echo.

git add -A
git commit -m "feat: Add auto-migration for time_schedule and update README

- Database migrations now run automatically on Railway deployment
- Changed scheduled_time/schedule_time to time_schedule across system
- Updated README with latest migration info
- Zero downtime migration with data preservation
- Improved consistency across all tables"

git push origin main

echo.
echo === Push completed! ===
echo Railway will automatically run migrations on deployment.
pause
