@echo off
echo ====================================
echo  EMERGENCY SEQUENCE STEPS FIX
echo ====================================

echo.
echo Step 1: Running emergency sequence data fix...
if not defined DATABASE_URL (
    echo ERROR: DATABASE_URL environment variable not set
    echo Please set your DATABASE_URL and try again
    pause
    exit /b 1
)

echo Running SQL fix...
psql "%DATABASE_URL%" -f emergency_sequence_fix.sql

if %ERRORLEVEL% EQU 0 (
    echo ✓ SQL fix completed successfully
) else (
    echo ✗ SQL fix failed - check error above
    pause
    exit /b 1
)

echo.
echo Step 2: Testing sequence API directly...
echo Making API call to check if steps are now visible...

curl -X GET "http://localhost:3000/api/sequences/394d567f-e5bd-476d-ae7c-c39f74819d70" -H "Content-Type: application/json" > sequence_test_result.json 2>nul

if exist sequence_test_result.json (
    echo ✓ API call completed - check sequence_test_result.json for results
    type sequence_test_result.json
) else (
    echo ⚠ Could not test API - make sure your app is running on port 3000
)

echo.
echo ====================================
echo  FIX COMPLETED!
echo ====================================
echo.
echo What was fixed:
echo - Set proper day/day_number values
echo - Fixed send_time from 'Invalid Date' to '10:00'
echo - Set message_type to 'text'
echo - Fixed created_at/updated_at timestamps
echo - Updated sequence step counts
echo.
echo Next steps:
echo 1. Restart your Go application if it's running
echo 2. Refresh your browser page
echo 3. Check if sequence steps now appear
echo.
pause
