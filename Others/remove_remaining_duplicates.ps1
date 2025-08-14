# Remove remaining duplicate functions
$filePath = "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\app.go"
$content = Get-Content $filePath -Raw

# Find and remove duplicate GetWorkerStatus at line 1705
Write-Host "Looking for duplicate GetWorkerStatus..."
$pattern1 = '(?s)\n// GetWorkerStatus[^\n]*\nfunc \(handler \*App\) GetWorkerStatus\(c \*fiber\.Ctx\) error \{[^}]+userRepo := repository\.GetUserRepository[^}]+// Get user[^}]+broadcastManager[^}]+return c\.JSON\(utils\.ResponseData[^}]+\}\s*\}'

if ($content -match $pattern1) {
    Write-Host "Found and removing duplicate GetWorkerStatus"
    $content = $content -replace $pattern1, ""
}

# Find and remove duplicate min at line 1786
Write-Host "Looking for duplicate min function..."
# Count occurrences
$minMatches = [regex]::Matches($content, 'func min\(a, b int\) int \{')
if ($minMatches.Count -gt 1) {
    Write-Host "Found $($minMatches.Count) min functions, removing duplicates"
    # Remove all but the first
    for ($i = $minMatches.Count - 1; $i -gt 0; $i--) {
        $start = $minMatches[$i].Index
        # Find the end of the function
        $end = $content.IndexOf("}", $content.IndexOf("}", $start) + 1) + 1
        $content = $content.Remove($start, $end - $start)
    }
}

# Find and remove duplicate countConnectedDevices at line 1794
Write-Host "Looking for duplicate countConnectedDevices..."
$countMatches = [regex]::Matches($content, 'func countConnectedDevices\(devices \[\]\*models\.UserDevice\) int \{')
if ($countMatches.Count -gt 1) {
    Write-Host "Found $($countMatches.Count) countConnectedDevices functions, removing duplicates"
    # Remove all but the first
    for ($i = $countMatches.Count - 1; $i -gt 0; $i--) {
        $start = $countMatches[$i].Index
        # Find the end of the function
        $endPos = $start
        $braceCount = 0
        $foundStart = $false
        
        for ($j = $start; $j -lt $content.Length; $j++) {
            if ($content[$j] -eq '{') {
                $braceCount++
                $foundStart = $true
            }
            elseif ($content[$j] -eq '}' -and $foundStart) {
                $braceCount--
                if ($braceCount -eq 0) {
                    $endPos = $j + 1
                    break
                }
            }
        }
        
        if ($endPos -gt $start) {
            $content = $content.Remove($start, $endPos - $start)
        }
    }
}

# Clean up any multiple empty lines
$content = $content -replace '\n\n\n+', "`n`n"

# Save the file
Set-Content $filePath -Value $content -NoNewline -Encoding UTF8

Write-Host "Removed duplicate functions!"
