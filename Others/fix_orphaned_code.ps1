# Fix orphaned code after removing duplicate function
$filePath = "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\app.go"
$content = Get-Content $filePath -Raw

# Find and remove the orphaned code that starts with userRepo := repository.GetUserRepository()
# This appears to be from line 1703 onwards until we hit a proper function declaration

# Use regex to find the orphaned code section
$pattern = '(?s)\n\s+userRepo := repository\.GetUserRepository\(\)[\s\S]*?(?=\n// [A-Z]|\nfunc |\z)'
$content = $content -replace $pattern, "`n"

# Write back
Set-Content $filePath -Value $content -NoNewline

Write-Host "Fixed orphaned code"
