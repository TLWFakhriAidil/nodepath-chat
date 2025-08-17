@echo off
echo ========================================
echo Pushing Sequence Detail Status Update to GitHub
echo ========================================

echo.
echo Changes made:
echo - Added two new status boxes: Failed and Remaining
echo - Now showing 5 main metric boxes at the top
echo - Each flow card now shows 4 stats: Should Send, Done Send, Failed Send, Remaining
echo - Timeline also updated to show all 4 statistics per flow
echo - Fetches contact data from sequence_contacts table to calculate real stats
echo - Status mapping: sent/completed = Done, failed = Failed, active/pending = Remaining
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "feat: Add failed and remaining status to sequence detail

- Add 'Contacts Failed Send Message' metric box (red)
- Add 'Contacts Remaining Send Message' metric box (yellow)
- Update flow cards to show 4 statistics each
- Fetch real contact data from /api/sequences/:id/contacts endpoint
- Calculate stats based on sequence_contacts status field
- Status mapping: sent/completed=Done, failed=Failed, active/pending=Remaining
- Update timeline to display all 4 metrics per flow"

REM Push to main branch
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause
