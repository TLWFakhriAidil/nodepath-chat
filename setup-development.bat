@echo off
echo ========================================
echo Configuring Application for Development
echo ========================================
echo.

echo [1/3] Creating development configuration...

REM Create a development env file that uses the existing remote database
(
echo # Development Configuration
echo APP_ENV=development
echo PORT=8080
echo.
echo # Using Remote Database (Same as Production)
echo MYSQL_URI=mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway
echo DATABASE_URL=admin_aqil:admin_aqil@tcp(157.245.206.124:3306^)/admin_railway?charset=utf8mb4^&parseTime=True^&loc=Local
echo.
echo # Legacy Vite Database Configuration
echo VITE_DB_HOST=157.245.206.124
echo VITE_DB_PORT=3306
echo VITE_DB_USER=admin_aqil
echo VITE_DB_PASSWORD=admin_aqil
echo VITE_DB_NAME=admin_railway
echo.
echo # Server URLs for local development
echo VITE_API_URL=http://localhost:8080/api
echo VITE_WS_URL=ws://localhost:8080/ws
echo API_URL=http://localhost:8080/api
echo.
echo # Development Mode Flag
echo DEVELOPMENT_MODE=true
) > .env.development

echo [2/3] Backing up current .env...
if not exist .env.backup (
    copy .env .env.backup >nul
    echo Original .env backed up as .env.backup
)

echo [3/3] Setting up for development...
copy .env.development .env >nul

echo.
echo ========================================
echo ✅ Development configuration complete!
echo ========================================
echo.
echo The application is now configured for development.
echo.
echo To start the application:
echo   go run cmd/server/main.go
echo.
echo Note: The database connection may fail if your IP is not whitelisted.
echo In that case, you can still use the fallback authentication:
echo.
echo   Email: admin@nodepath.com
echo   Password: admin123
echo.
echo To restore production configuration:
echo   copy .env.backup .env
echo.
pause
