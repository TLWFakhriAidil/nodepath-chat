@echo off
echo ========================================
echo UPDATING README DOCUMENTATION
echo ========================================

echo.
echo Committing README updates...
git add README.md STATUS_TARGETING_GUIDE.md
git commit -m "docs: Update README with lead management and status targeting features

- Added comprehensive lead management section
- Documented phone format changes (no + symbol)
- Explained niche support (single and comma-separated)
- Added status targeting documentation
- Updated database schema section
- Added import/export examples
- Included targeting examples and use cases"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================
echo README UPDATED!
echo ========================================
echo.
echo Documentation now includes:
echo 1. Lead management improvements
echo 2. Status targeting for campaigns/sequences
echo 3. Import/export format
echo 4. Targeting examples
echo 5. Database schema updates
echo.
pause
