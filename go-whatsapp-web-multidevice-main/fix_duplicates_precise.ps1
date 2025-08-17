# PowerShell script to remove duplicate functions by line ranges
$filePath = "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\app.go"
$lines = Get-Content $filePath

Write-Host "Total lines in file: $($lines.Count)"

# Create a list of line ranges to remove (0-based index)
# Based on the error messages:
# - GetSequenceSummary duplicate at line 1704 (index 1703)
# - GetWorkerStatus duplicate at line 1802 (index 1801)
# - min duplicate at line 1883 (index 1882)
# - max duplicate at line 1890 (index 1889)
# - countConnectedDevices duplicate at line 1897 (index 1896)
# - ResumeFailedWorkers duplicate at line 1909 (index 1908)

# We need to find the end of each function to remove the complete block
$linesToRemove = @()

# Function to find the end of a function starting at a given line
function Find-FunctionEnd {
    param($startLine, $lines)
    
    $braceCount = 0
    $inFunction = $false
    
    for ($i = $startLine; $i -lt $lines.Count; $i++) {
        if ($lines[$i] -match "^func\s+") {
            $inFunction = $true
        }
        
        if ($inFunction -or $i -eq $startLine) {
            $braceCount += ([regex]::Matches($lines[$i], "\{")).Count
            $braceCount -= ([regex]::Matches($lines[$i], "\}")).Count
            
            if ($braceCount -eq 0 -and $i -gt $startLine) {
                return $i
            }
        }
    }
    return $lines.Count - 1
}

# Find and mark duplicate functions for removal
Write-Host "Finding duplicate functions..."

# GetSequenceSummary at 1704 (but we need to start from the comment)
for ($i = 1700; $i -lt 1710; $i++) {
    if ($lines[$i] -match "// GetSequenceSummary gets sequence statistics") {
        $end = Find-FunctionEnd -startLine $i -lines $lines
        Write-Host "Found GetSequenceSummary duplicate from line $($i+1) to $($end+1)"
        $linesToRemove += @{Start = $i; End = $end}
        break
    }
}

# GetWorkerStatus at 1802
for ($i = 1798; $i -lt 1810; $i++) {
    if ($lines[$i] -match "// GetWorkerStatus gets the status of all device workers") {
        $end = Find-FunctionEnd -startLine $i -lines $lines
        Write-Host "Found GetWorkerStatus duplicate from line $($i+1) to $($end+1)"
        $linesToRemove += @{Start = $i; End = $end}
        break
    }
}

# min at 1883
for ($i = 1880; $i -lt 1890; $i++) {
    if ($lines[$i] -match "^func min\(") {
        $end = Find-FunctionEnd -startLine $i -lines $lines
        Write-Host "Found min duplicate from line $($i+1) to $($end+1)"
        $linesToRemove += @{Start = $i; End = $end}
        break
    }
}

# max at 1890
for ($i = 1887; $i -lt 1895; $i++) {
    if ($lines[$i] -match "^func max\(") {
        $end = Find-FunctionEnd -startLine $i -lines $lines
        Write-Host "Found max duplicate from line $($i+1) to $($end+1)"
        $linesToRemove += @{Start = $i; End = $end}
        break
    }
}

# countConnectedDevices at 1897
for ($i = 1894; $i -lt 1905; $i++) {
    if ($lines[$i] -match "^func countConnectedDevices\(") {
        $end = Find-FunctionEnd -startLine $i -lines $lines
        Write-Host "Found countConnectedDevices duplicate from line $($i+1) to $($end+1)"
        $linesToRemove += @{Start = $i; End = $end}
        break
    }
}

# ResumeFailedWorkers at 1909
for ($i = 1905; $i -lt 1915; $i++) {
    if ($lines[$i] -match "// ResumeFailedWorkers resumes all failed device workers") {
        $end = Find-FunctionEnd -startLine $i -lines $lines
        Write-Host "Found ResumeFailedWorkers duplicate from line $($i+1) to $($end+1)"
        $linesToRemove += @{Start = $i; End = $end}
        break
    }
}

# Sort ranges by start line in descending order to avoid index issues
$linesToRemove = $linesToRemove | Sort-Object -Property Start -Descending

# Create new content without the duplicate functions
$newLines = [System.Collections.ArrayList]::new($lines)

foreach ($range in $linesToRemove) {
    Write-Host "Removing lines $($range.Start + 1) to $($range.End + 1)"
    for ($i = $range.End; $i -ge $range.Start; $i--) {
        $newLines.RemoveAt($i)
    }
}

# Fix device.Name to device.DeviceName
Write-Host "Fixing device.Name to device.DeviceName..."
for ($i = 0; $i -lt $newLines.Count; $i++) {
    $newLines[$i] = $newLines[$i] -replace '(\s+)device\.Name(\s+|$|[^a-zA-Z])', '$1device.DeviceName$2'
}

# Fix unused 'user' variable - find and comment it out or remove the declaration
Write-Host "Fixing unused 'user' variable..."
for ($i = 1720; $i -lt 1730 -and $i -lt $newLines.Count; $i++) {
    if ($newLines[$i] -match "^\s*user, err := userRepo\.GetUserByID") {
        # Check if 'user' is used in the following lines
        $userUsed = $false
        for ($j = $i + 1; $j -lt ($i + 50) -and $j -lt $newLines.Count; $j++) {
            if ($newLines[$j] -match "\buser\." -or $newLines[$j] -match "\buser\b(?!\s*,\s*err)") {
                $userUsed = $true
                break
            }
        }
        if (-not $userUsed) {
            Write-Host "Found unused 'user' variable at line $($i+1), replacing with underscore"
            $newLines[$i] = $newLines[$i] -replace "user, err :=", "_, err :="
        }
        break
    }
}

# Save the fixed content
Write-Host "Saving fixed file..."
$newLines | Out-File -FilePath $filePath -Encoding UTF8

Write-Host "Done! File reduced from $($lines.Count) to $($newLines.Count) lines"
