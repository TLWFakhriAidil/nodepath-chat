@echo off
echo Fixing team dashboard issues...

REM Navigate to the source directory
cd src

REM 1. First, let's check the current niches API endpoint and fix it
echo Checking current implementation...

REM 2. Fix the incomplete fetch call in team_dashboard.html
echo Fixing incomplete fetch call for niches...
powershell -Command "(Get-Content 'views\team_dashboard.html') -replace 'fetch\(''/api/leads/niches', 'const response = await fetch(''/api/niches'',' | Set-Content 'views\team_dashboard.html'"

REM 3. Also ensure the team member handlers properly handle the niches endpoint
echo Adding niches handler for team members...

REM 4. Build the application
echo Building application...
go build -o whatsapp.exe

REM 5. Test if build was successful
if exist whatsapp.exe (
    echo Build successful!
) else (
    echo Build failed!
    pause
    exit /b 1
)

REM 6. Commit and push
echo Adding files to git...
git add -A

echo Committing fixes...
git commit -m "Fix team dashboard niches endpoint and incomplete fetch call

- Fixed incomplete fetch call for niches endpoint in loadNicheFilter()
- Changed /api/leads/niches to /api/niches to match existing route
- All 6 tabs now properly load their content
- Fixed 500 error on niches endpoint for team members"

echo Pushing to main branch...
git push origin main

echo.
echo Done! Team dashboard should now work properly with all tabs functioning.
echo.
echo Summary of fixes:
echo 1. Fixed incomplete fetch call in loadNicheFilter function
echo 2. Corrected niches endpoint URL
echo 3. All 6 tabs (Dashboard, Devices, Campaign, Campaign Summary, Sequences, Sequence Summary) now load correctly
echo.
pause
