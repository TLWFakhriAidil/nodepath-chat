@echo off
echo Fixing syntax errors in platform_sender.go...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src

:: Fix the malformed comments
powershell -Command "(Get-Content 'pkg\external\platform_sender.go') -replace 'imageURL// logrus', 'imageURL\n\t\t// logrus' | Set-Content 'pkg\external\platform_sender.go'"
powershell -Command "(Get-Content 'pkg\external\platform_sender.go') -replace '\}// logrus', '}\n\t// logrus' | Set-Content 'pkg\external\platform_sender.go'"
powershell -Command "(Get-Content 'pkg\external\platform_sender.go') -replace 'err\)// logrus', 'err)\n\t\t// logrus' | Set-Content 'pkg\external\platform_sender.go'"
powershell -Command "(Get-Content 'pkg\external\platform_sender.go') -replace '200\)\)', '200))' | Set-Content 'pkg\external\platform_sender.go'"

echo Syntax fixes applied!
