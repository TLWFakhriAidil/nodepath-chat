@echo off
echo ========================================
echo Pushing Inactive Count Calculation Fix to GitHub
echo ========================================

echo.
echo Changes made:
echo - Fixed inactive count calculation: inactive = total - active
echo - Removed switch statement that was incorrectly counting each status
echo - Now properly calculates: if total=2 and active=0, then inactive=2
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "fix: Correct inactive sequences calculation

- Changed logic to calculate inactive as: total sequences - active sequences
- Previously was counting each non-active status separately
- Now correctly shows inactive = total - active"

REM Push to main branch
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause
