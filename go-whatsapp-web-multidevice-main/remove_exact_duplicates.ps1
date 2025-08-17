# Remove specific duplicate functions by finding their exact locations
$filePath = "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\app.go"

# Read all lines
$lines = Get-Content $filePath

Write-Host "Total lines: $($lines.Count)"

# Track what to remove
$removeRanges = @()

# Find duplicate GetWorkerStatus (around line 1704)
for ($i = 1700; $i -lt 1710 -and $i -lt $lines.Count; $i++) {
    if ($lines[$i] -match "^// GetWorkerStatus gets the status of all device workers") {
        Write-Host "Found duplicate GetWorkerStatus at line $($i + 1)"
        # Find the end of this function
        $braceCount = 0
        $foundFunc = $false
        $endLine = $i
        
        for ($j = $i; $j -lt $lines.Count; $j++) {
            if ($lines[$j] -match "^func.*GetWorkerStatus") {
                $foundFunc = $true
            }
            if ($foundFunc) {
                $braceCount += ([regex]::Matches($lines[$j], "\{")).Count
                $braceCount -= ([regex]::Matches($lines[$j], "\}")).Count
                
                if ($braceCount -eq 0 -and $j -gt $i + 5) {
                    $endLine = $j
                    break
                }
            }
        }
        
        $removeRanges += @{Start = $i; End = $endLine}
        break
    }
}

# Find duplicate min function (around line 1786)
for ($i = 1780; $i -lt 1790 -and $i -lt $lines.Count; $i++) {
    if ($lines[$i] -match "^func min\(a, b int\) int \{") {
        Write-Host "Found duplicate min at line $($i + 1)"
        # Find the end (usually 5-6 lines)
        $endLine = $i
        for ($j = $i + 1; $j -lt $i + 10 -and $j -lt $lines.Count; $j++) {
            if ($lines[$j] -match "^\}$") {
                $endLine = $j
                break
            }
        }
        $removeRanges += @{Start = $i; End = $endLine}
        break
    }
}

# Find duplicate countConnectedDevices (around line 1794)
for ($i = 1790; $i -lt 1800 -and $i -lt $lines.Count; $i++) {
    if ($lines[$i] -match "^func countConnectedDevices\(devices \[\]\*models\.UserDevice\) int \{") {
        Write-Host "Found duplicate countConnectedDevices at line $($i + 1)"
        # Find the end
        $braceCount = 1
        $endLine = $i
        
        for ($j = $i + 1; $j -lt $lines.Count; $j++) {
            $braceCount += ([regex]::Matches($lines[$j], "\{")).Count
            $braceCount -= ([regex]::Matches($lines[$j], "\}")).Count
            
            if ($braceCount -eq 0) {
                $endLine = $j
                break
            }
        }
        
        $removeRanges += @{Start = $i; End = $endLine}
        break
    }
}

# Sort ranges by start line descending
$removeRanges = $removeRanges | Sort-Object -Property Start -Descending

# Create new content
$newLines = [System.Collections.ArrayList]::new()
$newLines.AddRange($lines)

# Remove the ranges
foreach ($range in $removeRanges) {
    Write-Host "Removing lines $($range.Start + 1) to $($range.End + 1)"
    $count = $range.End - $range.Start + 1
    $newLines.RemoveRange($range.Start, $count)
}

# Save the file
try {
    [System.IO.File]::WriteAllLines($filePath, $newLines, [System.Text.Encoding]::UTF8)
    Write-Host "Successfully saved file with $($newLines.Count) lines (removed $(($lines.Count - $newLines.Count)) lines)"
} catch {
    Write-Host "Error saving file: $_"
    Write-Host "Trying alternative method..."
    $newLines | Out-File $filePath -Encoding UTF8
}
