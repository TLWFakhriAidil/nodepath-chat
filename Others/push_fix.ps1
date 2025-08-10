# PowerShell script to commit and push
Write-Host "Adding all files..." -ForegroundColor Green
& 'C:\Program Files\Git\bin\git.exe' add -A

Write-Host "Committing changes..." -ForegroundColor Green
& 'C:\Program Files\Git\bin\git.exe' commit -m "Fix sequence device report - Fixed step_order column to use COALESCE(day_number, day, 1) - Added error logging for debugging - Dynamic step statistics display in frontend - Hide step statistics for regular campaigns"

Write-Host "Pushing to GitHub..." -ForegroundColor Green
& 'C:\Program Files\Git\bin\git.exe' push origin main

Write-Host "Done!" -ForegroundColor Green
