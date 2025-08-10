@echo off
echo ========================================
echo Device Management Update Script
echo ========================================
echo.
echo This script will update the dashboard to:
echo 1. Remove "Reset WhatsApp Session" tab
echo 2. Enhance Logout to remove session too
echo.
pause

echo.
echo Creating backup of dashboard.html...
copy "src\views\dashboard.html" "src\views\dashboard_backup_%date:~-4,4%%date:~-10,2%%date:~-7,2%_%time:~0,2%%time:~3,2%.html" >nul
echo Backup created!

echo.
echo Applying changes to dashboard.html...

REM PowerShell script to make the changes
powershell -Command "& {
    $file = 'src\views\dashboard.html'
    $content = Get-Content $file -Raw

    # Remove the Reset WhatsApp Session menu item
    $oldMenu = '<li><hr class=""dropdown-divider""></li>\s*<li><a class=""dropdown-item text-warning"" href=""#"" onclick=""resetDevice\(''(.+?)''\)"">\s*Reset WhatsApp Session\s*</a></li>'
    $newMenu = '<li><hr class=""dropdown-divider""></li>'
    $content = $content -replace $oldMenu, $newMenu

    # Update the logoutDevice function
    $oldLogout = 'function logoutDevice\(deviceId\) \{[\s\S]*?if \(confirm\(''Are you sure you want to logout this device from WhatsApp\?''\)\) \{[\s\S]*?\}\s*\}\s*\}'
    
    # Read new logout function from file
    $newLogoutContent = Get-Content 'new_logout_function.js' -Raw
    
    # Replace old function with new one
    if ($content -match $oldLogout) {
        $content = $content -replace $oldLogout, $newLogoutContent
        Write-Host 'Updated logoutDevice function' -ForegroundColor Green
    } else {
        Write-Host 'Could not find logoutDevice function to update' -ForegroundColor Yellow
    }

    # Remove the resetDevice function
    $resetPattern = '// Reset Device WhatsApp Session\s*function resetDevice\(deviceId\) \{[\s\S]*?\}\s*\}\s*\}'
    if ($content -match $resetPattern) {
        $content = $content -replace $resetPattern, ''
        Write-Host 'Removed resetDevice function' -ForegroundColor Green
    } else {
        Write-Host 'Could not find resetDevice function to remove' -ForegroundColor Yellow
    }

    # Save the updated content
    Set-Content $file $content -NoNewline
    Write-Host 'Dashboard.html updated successfully!' -ForegroundColor Green
}"

echo.
echo ========================================
echo Update Complete!
echo ========================================
echo.
echo Changes applied:
echo - Removed "Reset WhatsApp Session" from device menu
echo - Enhanced Logout to also remove session
echo - Removed redundant resetDevice function
echo.
echo Now when users click Logout:
echo - Device will be disconnected from WhatsApp
echo - Session will be completely removed
echo - They can scan QR code again to reconnect
echo.
pause