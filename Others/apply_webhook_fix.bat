@echo off
echo Applying webhook update to prevent duplicate devices...
echo.

REM Check if we're in the right directory
if not exist "src\ui\rest\webhook_lead.go" (
    echo ERROR: Cannot find src\ui\rest\webhook_lead.go
    echo Please run this script from the project root directory
    pause
    exit /b 1
)

REM Backup original files
echo Creating backups...
copy "src\ui\rest\webhook_lead.go" "src\ui\rest\webhook_lead_original.go" > nul
copy "src\repository\user_repository.go" "src\repository\user_repository_original.go" > nul

REM Apply webhook changes
echo Updating webhook logic...
copy /Y "src\ui\rest\webhook_lead_updated.go" "src\ui\rest\webhook_lead.go" > nul

REM Check if GetDeviceByUserAndName already exists
findstr /C:"GetDeviceByUserAndName" "src\repository\user_repository.go" > nul
if %ERRORLEVEL% NEQ 0 (
    echo Adding GetDeviceByUserAndName function to repository...
    echo. >> "src\repository\user_repository.go"
    type "src\repository\user_repository_addition.go" >> "src\repository\user_repository.go"
) else (
    echo GetDeviceByUserAndName function already exists in repository
)

REM Check if GetDB already exists
findstr /C:"func (r *UserRepository) GetDB()" "src\repository\user_repository.go" > nul
if %ERRORLEVEL% NEQ 0 (
    echo Adding GetDB helper function...
    echo. >> "src\repository\user_repository.go"
    echo // GetDB returns the database connection >> "src\repository\user_repository.go"
    echo func ^(r *UserRepository^) GetDB^(^) *sql.DB { >> "src\repository\user_repository.go"
    echo     return r.db >> "src\repository\user_repository.go"
    echo } >> "src\repository\user_repository.go"
) else (
    echo GetDB function already exists in repository
)

echo.
echo Changes applied successfully!
echo.
echo Summary of changes:
echo - Webhook now checks for existing devices by device_name first
echo - If device with same name exists, only updates the JID
echo - Prevents creating duplicate devices with similar names
echo - Added GetDeviceByUserAndName function to repository
echo.
echo Ready to build and test!
pause