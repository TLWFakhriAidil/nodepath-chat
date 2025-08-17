@echo off
echo ========================================
echo Pushing Sequence Detail Page Updates to GitHub
echo ========================================

echo.
echo Changes made:
echo - Removed breadcrumb, tabs (Analytics, Settings, Contacts)
echo - Added back button
echo - Added total leads count with trigger
echo - Changed layout to show Flow-based view
echo - Each flow shows total contacts should send and done send
echo - Timeline now shows flows with image, description, and counts
echo - Fixed SequenceDetailPage handler to include User data
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "feat: Redesign sequence detail page with flow-based view

- Remove breadcrumb navigation and unnecessary tabs
- Add back button for easier navigation  
- Display total leads count based on sequence trigger
- Implement flow-based cards showing should send/done send counts
- Update timeline to show flows with images and descriptions
- Fix SequenceDetailPage handler to pass user data to template
- Simplify UI to focus on sequence flow progress"

REM Push to main branch
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause
