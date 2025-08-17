@echo off
echo Fixing team authentication to use team_sessions table...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files...
git add -A

echo Committing authentication fixes...
git commit -m "Fix team authentication using team_sessions table

- Fixed syntax error in init_team_member.go (extra closing parenthesis)
- Added missing endpoints to isTeamAccessibleEndpoint function:
  - /api/campaigns (not just /api/campaigns/summary)
  - /api/sequences (not just /api/sequences/summary)
  - /api/analytics/dashboard
  - /api/leads/niches
  - /api/team-logout
- Team authentication now properly checks team_session cookie
- CustomAuth middleware now allows team members to access their endpoints
- Team sessions from database are now properly validated"

echo Pushing to main branch...
git push origin main

echo Done! Team authentication should now work properly.
pause