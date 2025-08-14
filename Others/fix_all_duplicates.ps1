# PowerShell script to fix all duplicate functions and errors
$filePath = "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\app.go"
$content = Get-Content $filePath -Raw

# First, let's identify all the duplicate functions and their line numbers
Write-Host "Analyzing file for duplicates..."

# Remove duplicate GetSequenceSummary (keep first one at line 1477, remove at 1704)
$pattern1 = '(?s)\n// GetSequenceSummary gets sequence statistics\nfunc \(handler \*App\) GetSequenceSummary\(c \*fiber\.Ctx\) error \{[^}]+userRepo := repository\.GetUserRepository\(\)[^}]+user, err := userRepo\.GetUserByID[^}]+sequences := \[]map\[string\]interface[^}]+return c\.JSON\(utils\.ResponseData[^}]+\}\s*\}'
if ($content -match $pattern1) {
    Write-Host "Found duplicate GetSequenceSummary, removing..."
    $content = $content -replace $pattern1, ""
}

# Remove duplicate GetWorkerStatus (keep first at 1547, remove at 1802)
$pattern2 = '(?s)\n// GetWorkerStatus gets the status of all device workers\nfunc \(handler \*App\) GetWorkerStatus\(c \*fiber\.Ctx\) error \{[^}]+Get user devices[^}]+broadcastManager := broadcast\.GetBroadcastManager[^}]+Build device worker info[^}]+device\.Name[^}]+return c\.JSON\(utils\.ResponseData[^}]+\}\s*\}'
if ($content -match $pattern2) {
    Write-Host "Found duplicate GetWorkerStatus, removing..."
    $content = $content -replace $pattern2, ""
}

# Remove duplicate min function (keep first at 1627, remove at 1883)
$pattern3 = '(?s)\n// Helper functions\nfunc min\(a, b int\) int \{[^}]+\}'
if ($content -match $pattern3) {
    Write-Host "Found duplicate min function, removing..."
    $content = $content -replace $pattern3, ""
}

# Remove duplicate max function (keep first at 1634, remove at 1890)
$pattern4 = '(?s)\nfunc max\(a, b int\) int \{[^}]+\}'
$matches = [regex]::Matches($content, $pattern4)
if ($matches.Count -gt 1) {
    Write-Host "Found duplicate max function, removing second occurrence..."
    # Remove the second occurrence
    $content = $content.Remove($matches[1].Index, $matches[1].Length)
}

# Remove duplicate countConnectedDevices (keep first at 1641, remove at 1897)
$pattern5 = '(?s)\nfunc countConnectedDevices\(devices \[\]\*models\.UserDevice\) int \{[^}]+\}'
$matches = [regex]::Matches($content, $pattern5)
if ($matches.Count -gt 1) {
    Write-Host "Found duplicate countConnectedDevices function, removing second occurrence..."
    # Remove the second occurrence
    $content = $content.Remove($matches[1].Index, $matches[1].Length)
}

# Remove duplicate ResumeFailedWorkers (keep first at 1653, remove at 1909)
$pattern6 = '(?s)\n// ResumeFailedWorkers resumes all failed device workers\nfunc \(handler \*App\) ResumeFailedWorkers\(c \*fiber\.Ctx\) error \{[^}]+Get all devices for user[^}]+Resume workers for devices[^}]+return c\.JSON\(utils\.ResponseData[^}]+\}\s*\}'
$matches = [regex]::Matches($content, $pattern6)
if ($matches.Count -gt 1) {
    Write-Host "Found duplicate ResumeFailedWorkers function, removing second occurrence..."
    # Start from the end to avoid index issues
    for ($i = $matches.Count - 1; $i -gt 0; $i--) {
        $content = $content.Remove($matches[$i].Index, $matches[$i].Length)
    }
}

# Fix device.Name to device.DeviceName
Write-Host "Fixing device.Name to device.DeviceName..."
$content = $content -replace 'device\.Name(?![a-zA-Z])', 'device.DeviceName'

# Fix the unused 'user' variable by removing the declaration if it's not used
Write-Host "Checking for unused 'user' variable..."
# This is trickier as we need context, so let's just report it for now

# Save the file
Write-Host "Saving fixed file..."
Set-Content $filePath -Value $content -NoNewline -Encoding UTF8

Write-Host "Done! All duplicate functions should be removed."
Write-Host "Note: You may still need to fix the unused 'user' variable at line 1725"
