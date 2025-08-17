# PowerShell script to fix duplicate function
$filePath = "src\ui\rest\app.go"
$lines = Get-Content $filePath

# Find the line numbers for both StopAllWorkers functions
$firstFunc = -1
$secondFunc = -1
$lineNum = 0

foreach ($line in $lines) {
    if ($line -match "^// StopAllWorkers stops all running device workers") {
        if ($firstFunc -eq -1) {
            $firstFunc = $lineNum
        } else {
            $secondFunc = $lineNum
        }
    }
    $lineNum++
}

Write-Host "First function at line: $($firstFunc + 1)"
Write-Host "Second function at line: $($secondFunc + 1)"

# Remove lines from first function until just before second function
# We need to remove from line 1702 to line 1747 (the end of first function)
$newLines = @()

for ($i = 0; $i -lt $lines.Length; $i++) {
    # Skip lines from the first StopAllWorkers function
    if ($i -ge 1702 -and $i -lt 1748) {
        continue
    }
    $newLines += $lines[$i]
}

# Write back
$newLines | Set-Content $filePath

Write-Host "Fixed! Removed duplicate function."
