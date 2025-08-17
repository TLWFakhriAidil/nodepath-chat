@echo off
echo Fixing all syntax errors in team dashboard...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

REM First backup current file
copy src\views\team_dashboard.html src\views\team_dashboard_backup_before_fix.html

echo Creating fixed version...

REM Create a Python script to fix all issues
echo import re > fix_team_dashboard.py
echo. >> fix_team_dashboard.py
echo with open('src/views/team_dashboard.html', 'r', encoding='utf-8') as f: >> fix_team_dashboard.py
echo     content = f.read() >> fix_team_dashboard.py
echo. >> fix_team_dashboard.py
echo # Fix missing newline after clearCampaignFilter >> fix_team_dashboard.py
echo content = content.replace('loadCampaignSummary();\n        }\n        // Load sequences', 'loadCampaignSummary();\n        }\n        \n        // Load sequences') >> fix_team_dashboard.py
echo. >> fix_team_dashboard.py
echo # Ensure all functions are properly formatted >> fix_team_dashboard.py
echo content = re.sub(r'}\n(\s*)//\s*([A-Z])', r'}\n\n\1// \2', content) >> fix_team_dashboard.py
echo. >> fix_team_dashboard.py
echo with open('src/views/team_dashboard.html', 'w', encoding='utf-8') as f: >> fix_team_dashboard.py
echo     f.write(content) >> fix_team_dashboard.py

python fix_team_dashboard.py

del fix_team_dashboard.py

echo Done fixing syntax errors
pause