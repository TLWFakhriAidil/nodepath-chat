# Fix unused device variable in app.go
$filePath = "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\app.go"
$content = Get-Content $filePath -Raw

# Replace the problematic loop in StopAllWorkers
$content = $content -replace 'for _, device := range devices \{', 'for range devices {'

# Write back
Set-Content $filePath -Value $content -NoNewline

Write-Host "Fixed unused device variable"
