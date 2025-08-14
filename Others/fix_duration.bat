@echo off
echo Fixing unused variable in platform_sender.go...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src

:: Fix the unused duration variable
powershell -Command "(Get-Content 'pkg\external\platform_sender.go') -replace 'duration := time.Since\(startTime\)', '_ = time.Since(startTime)' | Set-Content 'pkg\external\platform_sender.go.tmp'"
move /Y pkg\external\platform_sender.go.tmp pkg\external\platform_sender.go

echo Variable fix applied!
