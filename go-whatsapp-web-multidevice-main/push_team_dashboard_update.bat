@echo off
echo Committing and pushing team dashboard updates...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files...
git add -A

echo Committing changes...
git commit -m "Update team dashboard to exactly match master dashboard UI/UX

- Team dashboard now has exact same 6 tabs as master (Dashboard, Devices, Campaign, Campaign Summary, Sequences, Sequence Summary)
- Excluded AI Management and User Management tabs (admin-only)
- Same UI components: metrics cards, charts, calendar view, filters
- Read-only access for team members
- Data automatically filtered by device name matching username
- Complete feature parity with master dashboard for assigned devices"

echo Pushing to main branch...
git push origin main

echo Done!
pause