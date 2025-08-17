@echo off
echo ========================================
echo Hiding Worker/Redis Tabs & Updating Delete
echo ========================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Creating backup of dashboard.html...
copy "src\views\dashboard.html" "src\views\dashboard_backup_tabs_%date:~-4,4%%date:~-10,2%%date:~-7,2%_%time:~0,2%%time:~3,2%.html" >nul
echo Backup created!

echo.
echo Applying changes to dashboard.html...

REM PowerShell script to make the changes
powershell -Command "& {
    $file = 'src\views\dashboard.html'
    $content = Get-Content $file -Raw

    # 1. Hide Redis, Device Worker, All Worker buttons in the header
    # Comment out the system status icons group
    $oldButtons = '<!-- System Status Icons -->\s*<div class=""btn-group ms-3"" role=""group"">\s*<button.*?Redis.*?</button>\s*<button.*?Device Worker.*?</button>\s*<button.*?All Worker.*?</button>\s*</div>'
    $newButtons = '<!-- System Status Icons - HIDDEN -->
                <!-- <div class=""btn-group ms-3"" role=""group"">
                    <button class=""btn btn-sm btn-outline-secondary"" onclick=""checkRedis()"" title=""Check Redis Status"">
                        <i class=""bi bi-hdd-stack""></i> Redis
                    </button>
                    <button class=""btn btn-sm btn-outline-secondary"" onclick=""checkDeviceWorker()"" title=""Check Device Worker"">
                        <i class=""bi bi-cpu""></i> Device Worker
                    </button>
                    <button class=""btn btn-sm btn-outline-secondary"" onclick=""checkAllWorker()"" title=""Check All Workers"">
                        <i class=""bi bi-diagram-3""></i> All Worker
                    </button>
                </div> -->'
    
    if ($content -match $oldButtons) {
        $content = $content -replace $oldButtons, $newButtons
        Write-Host 'Hidden system status buttons' -ForegroundColor Green
    }

    # 2. Hide Worker Status tab
    $oldWorkerTab = '<li class=""nav-item"" role=""presentation"">\s*<button class=""nav-link"" id=""worker-status-tab"".*?>\s*<i class=""bi bi-cpu""></i>\s*Worker Status\s*</button>\s*</li>'
    $newWorkerTab = '<!-- Worker Status Tab - HIDDEN -->
                <!-- <li class=""nav-item"" role=""presentation"">
                    <button class=""nav-link"" id=""worker-status-tab"" data-bs-toggle=""tab"" data-bs-target=""#worker-status"" type=""button"">
                        <i class=""bi bi-cpu""></i> Worker Status
                    </button>
                </li> -->'
    
    if ($content -match $oldWorkerTab) {
        $content = $content -replace $oldWorkerTab, $newWorkerTab
        Write-Host 'Hidden Worker Status tab' -ForegroundColor Green
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
echo - Hidden Redis button
echo - Hidden Device Worker button  
echo - Hidden All Worker button
echo - Hidden Worker Status tab
echo - Delete device now uses SweetAlert2 (next step)
echo.
pause