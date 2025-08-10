@echo off
echo Updating dashboard.html...

REM Create backup
copy src\views\dashboard.html src\views\dashboard_backup_%date:~-4,4%%date:~-10,2%%date:~-7,2%.html

REM Use PowerShell to make the changes
powershell -Command "& {
    # Read the file
    $content = Get-Content -Path 'src\views\dashboard.html' -Raw
    
    # Remove Reset WhatsApp Session menu item
    $pattern1 = '<li><a class=""dropdown-item text-warning"" href=""#"" onclick=""resetDevice[^<]+<\/a><\/li>'
    $content = $content -replace $pattern1, ''
    
    # Also remove the icon line if present
    $pattern2 = '<i class=""bi bi-arrow-counterclockwise me-2""><\/i>Reset WhatsApp Session'
    $content = $content -replace $pattern2, ''
    
    # Save the file
    $content | Set-Content -Path 'src\views\dashboard.html' -NoNewline
}"

echo Dashboard updated successfully!
echo.
echo IMPORTANT: You still need to manually update the logoutDevice function
echo to include the session reset functionality.
echo.
pause
