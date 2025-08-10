@echo off
echo ========================================
echo Pushing Sequence Summary Updates to GitHub
echo ========================================

echo.
echo Changes made:
echo - Removed "Draft" section from sequence summary
echo - Changed "Paused" to "Inactive" in the summary cards
echo - Added "Total Devices" column to recent sequences table
echo - Backend now includes total_devices count for sequences
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "feat: Update sequence summary UI and add device count

- Remove 'Draft' status from sequence summary cards
- Change 'Paused' to 'Inactive' for clearer status indication
- Add 'Total Devices' column to recent sequences table
- Update backend to include total device count for each sequence
- Since sequences use all user devices, show user's total device count"

REM Push to main branch
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause
