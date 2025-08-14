@echo off
echo Fixing Sequence Device Report calculation logic...

REM Add changes
git add src/ui/rest/app.go

REM Commit with descriptive message
git commit -m "Fix sequence device report calculation to match summary page logic

- Changed shouldSend calculation to match summary page: done + failed + remaining
- Updated query to get remaining_send count separately
- Fixed overall totals calculation to use same logic as summary
- Added debug logging for better troubleshooting
- Ensures device report shows consistent numbers with sequence summary"

REM Push to main
git push origin main

echo.
echo Done! Check GitHub Actions for deployment status.
pause