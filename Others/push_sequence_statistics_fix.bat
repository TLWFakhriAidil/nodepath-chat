@echo off
echo ========================================
echo Pushing Sequence Summary Statistics Fix to GitHub
echo ========================================

echo.
echo Changes made:
echo - Fixed "undefined" display for inactive sequences count
echo - Added total_devices to the summary statistics
echo - Added total_success and total_remaining to contact statistics
echo - Updated frontend to show 4 statistics: Total Devices, Total Contacts, Success, Remaining
echo - Updated recent sequences table to show Success and Remaining columns
echo - Temporary workaround: Success shows as 0 until backend models are updated
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "fix: Sequence summary statistics and undefined inactive count

- Fix undefined display for inactive sequences count by using default case
- Add total_devices to summary response
- Add total_success and total_remaining statistics
- Update frontend Contact Statistics to show 4 metrics
- Update recent sequences table with Success and Remaining columns
- Note: Success counts show as 0 until SequenceResponse model is updated with CompletedContacts field"

REM Push to main branch
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause
