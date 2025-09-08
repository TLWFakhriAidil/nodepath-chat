@echo off
echo ========================================
echo Setting up Local MySQL Database
echo ========================================
echo.

REM Check if Docker is installed
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Docker is not installed or not in PATH
    echo Please install Docker Desktop from https://www.docker.com/products/docker-desktop
    pause
    exit /b 1
)

echo [1/5] Checking if MySQL container already exists...
docker ps -a | findstr mysql-nodepath >nul 2>&1
if %errorlevel% equ 0 (
    echo Removing existing MySQL container...
    docker stop mysql-nodepath >nul 2>&1
    docker rm mysql-nodepath >nul 2>&1
)

echo [2/5] Starting MySQL database in Docker...
docker run -d ^
    --name mysql-nodepath ^
    -e MYSQL_ROOT_PASSWORD=admin_aqil ^
    -e MYSQL_DATABASE=admin_railway ^
    -e MYSQL_USER=admin_aqil ^
    -e MYSQL_PASSWORD=admin_aqil ^
    -p 3306:3306 ^
    mysql:8.0

if %errorlevel% neq 0 (
    echo ERROR: Failed to start MySQL container
    pause
    exit /b 1
)

echo [3/5] Waiting for MySQL to be ready...
timeout /t 10 /nobreak >nul

echo [4/5] Creating .env.local file...
(
echo # Local Development Database Configuration
echo APP_ENV=development
echo PORT=8080
echo.
echo # Local MySQL Database
echo MYSQL_URI=mysql://admin_aqil:admin_aqil@localhost:3306/admin_railway
echo DATABASE_URL=admin_aqil:admin_aqil@tcp(localhost:3306^)/admin_railway?charset=utf8mb4^&parseTime=True^&loc=Local
echo.
echo # Legacy Vite Database Configuration
echo VITE_DB_HOST=localhost
echo VITE_DB_PORT=3306
echo VITE_DB_USER=admin_aqil
echo VITE_DB_PASSWORD=admin_aqil
echo VITE_DB_NAME=admin_railway
echo.
echo # Server URLs
echo VITE_API_URL=http://localhost:8080/api
echo VITE_WS_URL=ws://localhost:8080/ws
echo API_URL=http://localhost:8080/api
) > .env.local

echo [5/5] Backup original .env and use local version...
if not exist .env.original (
    copy .env .env.original >nul
)
copy .env.local .env >nul

echo.
echo ========================================
echo ✅ Local MySQL database setup complete!
echo ========================================
echo.
echo Database Details:
echo   Host: localhost
echo   Port: 3306
echo   Database: admin_railway
echo   Username: admin_aqil
echo   Password: admin_aqil
echo.
echo The .env file has been updated for local development.
echo Original .env backed up as .env.original
echo.
echo To start the application:
echo   go run cmd/server/main.go
echo.
echo To restore production .env:
echo   copy .env.original .env
echo.
pause
