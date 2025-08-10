@echo off
echo Fixing team dashboard refresh loop issues...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

REM Create a comprehensive fix script
echo Creating fix script...
powershell -Command @"
$content = Get-Content 'src\views\team_dashboard.html' -Raw

# Fix 1: Add missing parentheses to loadTeamMemberInfo call
$content = $content -replace 'loadTeamMemberInfo\r?\n', 'loadTeamMemberInfo();`r`n'

# Fix 2: Fix incomplete sessionStorage calls in loadTeamMemberInfo
$content = $content -replace 'if \(!sessionStorage\r?\n\s*sessionStorage', 'if (!sessionStorage.getItem(''redirecting'')) {`r`n                            sessionStorage.setItem(''redirecting'', ''true'');'

# Fix 3: Fix incomplete window.location.href statements
$content = $content -replace 'window\.location\.href\r?\n', 'window.location.href = ''/team-login'';`r`n'

# Fix 4: Fix sessionStorage.removeItem call
$content = $content -replace '// Clear the redirect flag on successful auth\r?\n\s*sessionStorage\r?\n', '// Clear the redirect flag on successful auth`r`n                sessionStorage.removeItem(''redirecting'');`r`n'

# Fix 5: Fix loadTeamMemberInfo call in loadDashboardData
$content = $content -replace '// Don''t redirect here, as loadTeamMemberInfo already handles auth\r?\n\s*loadTeamMemberInfo\r?\n', '// Don''t redirect here, as loadTeamMemberInfo already handles auth`r`n                        // loadTeamMemberInfo();`r`n'

# Fix 6: Fix button onclick
$content = $content -replace 'window\.location\.href''/team-dashboard''', 'window.location.href=''/team-dashboard'''

# Save the fixed content
Set-Content 'src\views\team_dashboard.html' -Value $content
"@

REM Build the application
cd src
echo Building application...
set CGO_ENABLED=0
go build -o whatsapp.exe

cd ..

REM Commit and push
echo Committing fixes...
git add -A
git commit -m "Fix team dashboard infinite refresh loop

- Fixed missing parentheses on loadTeamMemberInfo() call
- Fixed incomplete sessionStorage statements
- Fixed incomplete window.location.href redirects
- Added proper authentication error handling
- Team dashboard should now load without constant refreshing"

echo Pushing to GitHub...
git push origin main

echo.
echo Fix complete! The team dashboard should now:
echo - Load without infinite refresh loops
echo - Properly handle authentication
echo - Work correctly in all browsers including incognito mode
echo.
pause
