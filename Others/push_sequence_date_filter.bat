@echo off
echo ========================================
echo Pushing Sequence Date Filter and Logic Update to GitHub
echo ========================================

echo.
echo Changes made:
echo - Added date filter section with start/end date inputs
echo - Filter applies to Done Send and Failed Send based on completed_at column
echo - Remaining count adjusts based on filtered results
echo - Updated calculation logic:
echo   - Should Send: leads where trigger matches sequence trigger
echo   - Done Send: sequence_contacts where status='sent'
echo   - Failed Send: sequence_contacts where status='failed'
echo   - Remaining: Should Send - Done Send - Failed Send
echo - Per-flow statistics use sequence_stepid to group contacts
echo - Added comprehensive documentation in README
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "feat: Add date filter and improve sequence progress logic

- Add date range filter for Done Send and Failed Send statistics
- Filter based on completed_at column in sequence_contacts table
- Implement proper calculation logic:
  * Should Send: count leads matching sequence trigger
  * Done/Failed: count sequence_contacts by status
  * Remaining: calculated difference
- Per-flow stats grouped by sequence_stepid column
- Store unfiltered data to maintain accurate remaining count
- Update README with detailed progress tracking documentation"

REM Push to main branch
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause
