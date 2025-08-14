@echo off
echo ========================================
echo Pushing Sequence Detail Box Layout Update to GitHub
echo ========================================

echo.
echo Changes made:
echo - Redesigned page with metric card boxes layout
echo - First box: Total Flows
echo - Second box: Total Contacts Should Send
echo - Third box: Contacts Done Send Message
echo - Additional boxes for each flow showing should/done counts
echo - Removed Edit, Start, Delete buttons
echo - Updated back button to redirect to sequences page
echo - Improved timeline design with clear flow information
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "feat: Redesign sequence detail with metric boxes layout

- Replace previous layout with clean metric card boxes
- Add 3 main metric boxes: Total Flows, Should Send, Done Send
- Create individual flow boxes showing per-flow statistics
- Remove action buttons (Edit, Start, Delete)
- Update back button to properly redirect to sequences page
- Enhance timeline with better visual hierarchy and stats
- Implement responsive grid layout for flow cards"

REM Push to main branch
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause
