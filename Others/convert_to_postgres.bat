@echo off
echo Converting all SQLite placeholders to PostgreSQL...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src

echo.
echo Fixing repository files...

REM Fix broadcast_repository.go
powershell -Command "(Get-Content repository\broadcast_repository.go) -replace 'VALUES \(\?, \?, \?, \?, \?, \?, \?, \?, \?, \?, \?, \?\)', 'VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)' | Set-Content repository\broadcast_repository.go"
powershell -Command "(Get-Content repository\broadcast_repository.go) -replace 'SET status = \?, processed_at = \?', 'SET status = $1, processed_at = $2' | Set-Content repository\broadcast_repository.go"
powershell -Command "(Get-Content repository\broadcast_repository.go) -replace 'WHERE id = \?', 'WHERE id = $1' | Set-Content repository\broadcast_repository.go"
powershell -Command "(Get-Content repository\broadcast_repository.go) -replace 'SET error_message = \?', 'SET error_message = $1' | Set-Content repository\broadcast_repository.go"
powershell -Command "(Get-Content repository\broadcast_repository.go) -replace 'VALUES \(\?, \?, \?, \?, \?, ''running'', \?\)', 'VALUES ($1, $2, $3, $4, $5, ''running'', $6)' | Set-Content repository\broadcast_repository.go"
powershell -Command "(Get-Content repository\broadcast_repository.go) -replace 'SET processed_contacts = \?, successful_contacts = \?, failed_contacts = \?', 'SET processed_contacts = $1, successful_contacts = $2, failed_contacts = $3' | Set-Content repository\broadcast_repository.go"
powershell -Command "(Get-Content repository\broadcast_repository.go) -replace 'SET status = ''completed'', completed_at = \?', 'SET status = ''completed'', completed_at = $1' | Set-Content repository\broadcast_repository.go"
powershell -Command "(Get-Content repository\broadcast_repository.go) -replace 'scheduled_at \<= \?\)', 'scheduled_at <= $1)' | Set-Content repository\broadcast_repository.go"
powershell -Command "(Get-Content repository\broadcast_repository.go) -replace 'LIMIT \?', 'LIMIT $2' | Set-Content repository\broadcast_repository.go"

REM Fix campaign_repository.go
powershell -Command "(Get-Content repository\campaign_repository.go) -replace 'VALUES \(\?, \?, \?, \?, \?, \?, \?, \?, \?, \?, \?, \?\)', 'VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)' | Set-Content repository\campaign_repository.go"
powershell -Command "(Get-Content repository\campaign_repository.go) -replace 'WHERE scheduled_date = \?', 'WHERE scheduled_date = $1' | Set-Content repository\campaign_repository.go"
powershell -Command "(Get-Content repository\campaign_repository.go) -replace 'SET title = \?, niche = \?, message = \?, image_url = \?,', 'SET title = $1, niche = $2, message = $3, image_url = $4,' | Set-Content repository\campaign_repository.go"
powershell -Command "(Get-Content repository\campaign_repository.go) -replace 'scheduled_date = \?, scheduled_time = \?, status = \?, updated_at = \?', 'scheduled_date = $5, scheduled_time = $6, status = $7, updated_at = $8' | Set-Content repository\campaign_repository.go"
powershell -Command "(Get-Content repository\campaign_repository.go) -replace 'WHERE id = \?', 'WHERE id = $9' | Set-Content repository\campaign_repository.go"

echo.
echo Repository files converted!
cd ..
