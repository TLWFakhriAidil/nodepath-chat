#!/bin/bash
# Auto-setup script for WhatsApp Multi-Device on Railway
# This script configures all environment variables automatically

echo "==================================="
echo "WhatsApp Multi-Device Auto Setup"
echo "==================================="

# Check if railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo "❌ Railway CLI not found. Installing..."
    echo "Please install Railway CLI from: https://docs.railway.app/develop/cli"
    exit 1
fi

echo "✅ Railway CLI found"

# Login to Railway
echo "📝 Please login to Railway..."
railway login

# Link to project
echo "🔗 Linking to Railway project..."
railway link

echo "⚙️ Setting up environment variables..."

# Core Database Configuration
echo "1️⃣ Setting up database configuration..."
railway variables set DB_URI="$DATABASE_URL"

# Application Configuration
echo "2️⃣ Setting up application configuration..."
railway variables set APP_PORT=3000
railway variables set APP_DEBUG=false
railway variables set APP_OS="WhatsApp Business System"
railway variables set APP_BASIC_AUTH="admin:changeme123"
railway variables set APP_CHAT_FLUSH_INTERVAL=30

# WhatsApp Features
echo "3️⃣ Setting up WhatsApp features..."
railway variables set WHATSAPP_CHAT_STORAGE=true
railway variables set WHATSAPP_ACCOUNT_VALIDATION=true
railway variables set WHATSAPP_AUTO_REPLY="Thank you for contacting us. We will respond shortly."

# Optional: Webhook Configuration
echo "4️⃣ Setting up webhook (optional)..."
read -p "Do you want to configure webhook? (y/n): " setup_webhook
if [ "$setup_webhook" = "y" ]; then
    read -p "Enter webhook URL: " webhook_url
    read -p "Enter webhook secret: " webhook_secret
    railway variables set WHATSAPP_WEBHOOK="$webhook_url"
    railway variables set WHATSAPP_WEBHOOK_SECRET="$webhook_secret"
fi

# Additional Performance Settings
echo "5️⃣ Setting up performance optimizations..."
railway variables set NODE_ENV=production
railway variables set NODE_TLS_REJECT_UNAUTHORIZED=0

echo ""
echo "✅ Environment variables configured!"
echo ""

# Deploy
echo "🚀 Deploying to Railway..."
railway up

echo ""
echo "==================================="
echo "✅ Setup Complete!"
echo "==================================="
echo ""
echo "Your WhatsApp Multi-Device system is now configured with:"
echo "- Database connection"
echo "- Chat storage enabled"
echo "- Auto-reply enabled"
echo "- Account validation enabled"
echo "- Admin login: admin@whatsapp.com / changeme123"
echo ""
echo "🌐 Your app will be available at your Railway domain"
echo ""
