@echo off
echo ========================================
echo Pushing Sequence Summary UI Update to GitHub
echo ========================================

echo.
echo Changes made:
echo - Added 6 metric boxes in Sequence Summary
echo - Removed Contact Statistics section
echo - Changed "Recent Sequences" to "Detail Sequences"
echo - Added Trigger column to table
echo - Added comprehensive statistics columns:
echo   - Total Flows
echo   - Total Contacts Should Send
echo   - Contacts Done Send Message
echo   - Contacts Failed Send Message
echo   - Contacts Remaining Send Message
echo - Updated backend to calculate all statistics
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "feat: Enhance sequence summary with detailed statistics

- Add 6 metric boxes: Total Sequences, Total Flows, Should Send, Done Send, Failed Send, Remaining
- Remove Contact Statistics section for cleaner UI
- Rename 'Recent Sequences' to 'Detail Sequences'
- Add Trigger column to sequences table
- Add comprehensive statistics columns for each sequence
- Update backend to calculate statistics from sequence_contacts table
- Track sent/failed status and calculate remaining dynamically"

REM Push to main branch
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause
