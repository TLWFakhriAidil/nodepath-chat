@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository

REM Fix the error handling to not ignore the UPDATE failure
powershell -Command "(Get-Content broadcast_repository.go) -replace 'logrus.Errorf\(\"Failed to update message status: %%v\", err\)', 'logrus.Errorf(\"Failed to update message status: %%v\", err); return nil, fmt.Errorf(\"failed to update processing_worker_id: %%w\", err)' | Set-Content broadcast_repository.go"

echo Fixed error handling for UPDATE failure