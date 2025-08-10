@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Add History Triggers and Customer Name bulk update" -m "Frontend changes:" -m "- Added 'History Triggers' filter to show historical triggers from broadcast_messages" -m "- Added Customer Name field to bulk update modal" -m "- History filter loads triggers from sequence_steps via broadcast_messages" -m "" -m "Backend changes:" -m "- GetDeviceLeads now accepts include_history=true parameter" -m "- When history is requested, queries broadcast_messages for historical triggers" -m "- UpdateLead now also updates recipient_name in broadcast_messages table" -m "- Bulk name update will update both leads and broadcast_messages tables" -m "" -m "Functionality:" -m "- History Triggers shows all triggers a lead has ever had (current + historical)" -m "- Customer name bulk update updates name in leads table and recipient_name in broadcast_messages"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
