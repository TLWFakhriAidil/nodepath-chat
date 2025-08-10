@echo off
echo ===================================
echo WhatsApp Multi-Device Auto Setup
echo ===================================

REM Check if railway CLI is installed
where railway >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Railway CLI not found. Installing...
    echo Please install Railway CLI from: https://docs.railway.app/develop/cli
    pause
    exit /b 1
)

echo Railway CLI found

REM Login to Railway
echo Please login to Railway...
railway login

REM Link to project
echo Linking to Railway project...
railway link

echo Setting up environment variables...

REM Core Database Configuration
echo 1. Setting up database configuration...
railway variables set DB_URI="$DATABASE_URL"

REM Application Configuration
echo 2. Setting up application configuration...
railway variables set APP_PORT=3000
railway variables set APP_DEBUG=false
railway variables set APP_OS="WhatsApp Business System"
railway variables set APP_BASIC_AUTH="admin:changeme123"
railway variables set APP_CHAT_FLUSH_INTERVAL=30

REM WhatsApp Features
echo 3. Setting up WhatsApp features...
railway variables set WHATSAPP_CHAT_STORAGE=true
railway variables set WHATSAPP_ACCOUNT_VALIDATION=true
railway variables set WHATSAPP_AUTO_REPLY="Thank you for contacting us. We will respond shortly."

REM Optional: Webhook Configuration
echo 4. Setting up webhook (optional)...
set /p setup_webhook="Do you want to configure webhook? (y/n): "
if "%setup_webhook%"=="y" (
    set /p webhook_url="Enter webhook URL: "
    set /p webhook_secret="Enter webhook secret: "
    railway variables set WHATSAPP_WEBHOOK="%webhook_url%"
    railway variables set WHATSAPP_WEBHOOK_SECRET="%webhook_secret%"
)

REM Additional Performance Settings
echo 5. Setting up performance optimizations...
railway variables set NODE_ENV=production
railway variables set NODE_TLS_REJECT_UNAUTHORIZED=0

echo.
echo Environment variables configured!
echo.

REM Deploy
echo Deploying to Railway...
railway up

echo.
echo ===================================
echo Setup Complete!
echo ===================================
echo.
echo Your WhatsApp Multi-Device system is now configured with:
echo - Database connection
echo - Chat storage enabled
echo - Auto-reply enabled
echo - Account validation enabled
echo - Admin login: admin@whatsapp.com / changeme123
echo.
echo Your app will be available at your Railway domain
echo.
pause
