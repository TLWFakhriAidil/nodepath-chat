@echo off
echo Fixing sequence column names...
echo ==============================

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"

echo.
echo Fixing repository/sequence_repository.go...
powershell -Command "(Get-Content repository/sequence_repository.go) -replace 'added_at', 'completed_at' | Set-Content repository/sequence_repository.go"
powershell -Command "(Get-Content repository/sequence_repository.go) -replace 'current_day', 'current_step' | Set-Content repository/sequence_repository.go"

echo.
echo Fixing usecase/sequence_trigger_processor.go...
powershell -Command "(Get-Content usecase/sequence_trigger_processor.go) -replace 'current_day', 'current_step' | Set-Content usecase/sequence_trigger_processor.go"
powershell -Command "(Get-Content usecase/sequence_trigger_processor.go) -replace 'enrolled_at', 'completed_at' | Set-Content usecase/sequence_trigger_processor.go"

echo.
echo Fixing models/sequence.go...
powershell -Command "(Get-Content models/sequence.go) -replace 'added_at', 'completed_at' | Set-Content models/sequence.go"

echo.
echo Fixing domains/sequence/sequence.go...
powershell -Command "(Get-Content domains/sequence/sequence.go) -replace 'added_at', 'completed_at' | Set-Content domains/sequence/sequence.go"

echo.
echo Fixing database/connection.go...
powershell -Command "(Get-Content database/connection.go) -replace 'added_at', 'completed_at' | Set-Content database/connection.go"
powershell -Command "(Get-Content database/connection.go) -replace 'current_day', 'current_step' | Set-Content database/connection.go"

echo.
echo All fixes applied!
echo.
pause
