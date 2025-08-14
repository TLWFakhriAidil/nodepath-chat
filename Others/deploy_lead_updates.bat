@echo off
echo ========================================
echo DEPLOYING LEAD MANAGEMENT UPDATES
echo ========================================

echo.
echo Committing lead management improvements...
git add -A
git commit -m "feat: Improve lead management system

- Phone number format: Remove + from placeholder (60123456789)
- Niche field: Simplified label and support for multiple niches (EXSTART,ITADRESS)
- Status options: Changed to prospect/customer only
- Journey field: Renamed to Additional Note
- Added niche filtering: Dynamic filters based on unique niches
- Filter system: Both status and niche filters work together
- Import/Export: Updated CSV format with new field names
- Default status: New leads default to 'prospect'

UI improvements:
- Cleaner form layout
- Better placeholder examples
- Dynamic niche filter generation
- Improved user experience"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================
echo DEPLOYMENT COMPLETE!
echo ========================================
echo.
echo Lead Management Updates:
echo 1. Phone: No + symbol (60123456789)
echo 2. Niche: Single or multiple (EXSTART,ITADRESS)
echo 3. Status: prospect/customer
echo 4. Additional Note instead of Journey
echo 5. Dynamic niche filters
echo.
pause
