@echo off
echo Fixing team dashboard niches endpoint...

cd src

REM Fix the niches endpoint URL to match the actual route
echo Updating niches endpoint URL from /api/leads/niches to /api/niches...
powershell -Command "(Get-Content 'views\team_dashboard.html') -replace '/api/leads/niches', '/api/niches' | Set-Content 'views\team_dashboard.html'"

REM Build the application
echo Building application without CGO...
set CGO_ENABLED=0
go build -o whatsapp.exe

REM Check if build was successful
if exist whatsapp.exe (
    echo Build successful!
) else (
    echo Build failed!
    pause
    exit /b 1
)

REM Commit and push changes
echo Committing changes...
git add -A
git commit -m "Fix team dashboard niches endpoint URL

- Changed /api/leads/niches to /api/niches to match actual route
- This fixes the 500 error when loading the team dashboard
- All 6 tabs should now work properly"

echo Pushing to GitHub...
git push origin main

echo.
echo Fix complete! The team dashboard should now:
echo - Load without 500 errors
echo - Show all 6 tabs working properly
echo - Allow team members to view their assigned data
echo.
pause
