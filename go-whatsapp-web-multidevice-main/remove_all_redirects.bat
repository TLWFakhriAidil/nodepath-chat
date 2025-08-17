@echo off
echo Removing ALL authentication redirects from team dashboard...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

powershell -Command @"
$content = Get-Content 'src\views\team_dashboard.html' -Raw

# Remove ALL 401 redirect blocks
$content = $content -replace '(?s)if \(response\.status === 401\) \{[^}]*?window\.location\.href = ''/team-login''[^}]*?\}', 'if (response.status === 401) { console.error(''401 error but not redirecting''); return; }'

# But keep the logout redirect
$content = $content -replace '\.then\(\(\) => \{[\r\n\s]*window\.location\.href = ''/team-login'';[\r\n\s]*\}\)', '.then(() => { window.location.href = ''/team-login''; })'

Set-Content 'src\views\team_dashboard.html' -Value $content -NoNewline
"@

echo Building...
cd src
set CGO_ENABLED=0
go build -o whatsapp.exe
cd ..

echo Committing...
git add -A
git commit -m "Remove ALL auth redirects from team dashboard"
git push origin main

echo Done!
pause
