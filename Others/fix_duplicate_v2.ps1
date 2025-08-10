# PowerShell script to fix duplicate function
$filePath = "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\app.go"
$lines = Get-Content $filePath

# Find lines with StopAllWorkers
$lineNumbers = @()
for ($i = 0; $i -lt $lines.Length; $i++) {
    if ($lines[$i] -match "^// StopAllWorkers stops all running device workers") {
        $lineNumbers += $i
        Write-Host "Found StopAllWorkers at line $($i + 1)"
    }
}

if ($lineNumbers.Count -eq 2) {
    Write-Host "Found duplicate functions. Removing first one..."
    
    # Find the end of the first function
    $startLine = $lineNumbers[0]
    $endLine = $startLine
    $braceCount = 0
    $inFunction = $false
    
    for ($i = $startLine; $i -lt $lineNumbers[1]; $i++) {
        if ($lines[$i] -match "^func.*StopAllWorkers") {
            $inFunction = $true
        }
        
        if ($inFunction) {
            if ($lines[$i] -match "\{") {
                $braceCount += ($lines[$i] -split "\{").Count - 1
            }
            if ($lines[$i] -match "\}") {
                $braceCount -= ($lines[$i] -split "\}").Count - 1
                if ($braceCount -eq 0) {
                    $endLine = $i
                    break
                }
            }
        }
    }
    
    Write-Host "Removing lines from $($startLine + 1) to $($endLine + 1)"
    
    # Create new content without the duplicate
    $newLines = @()
    for ($i = 0; $i -lt $lines.Length; $i++) {
        if ($i -lt $startLine -or $i -gt $endLine) {
            $newLines += $lines[$i]
        }
    }
    
    # Write back
    $newLines | Out-File -FilePath $filePath -Encoding UTF8
    Write-Host "Fixed! Removed duplicate function."
} else {
    Write-Host "Did not find exactly 2 StopAllWorkers functions. Found: $($lineNumbers.Count)"
}
