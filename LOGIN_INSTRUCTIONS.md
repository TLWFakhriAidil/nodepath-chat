# Login Instructions for NodePath Chat (Local Development)

## Current Status
The application is running in **fallback mode** because the database connection is blocked from your local IP address. This is normal for local development.

## How to Login

### Use one of these test credentials:

#### Admin Account
- **Email:** `admin@nodepath.com`
- **Password:** `admin123`

#### Test Account
- **Email:** `test@nodepath.com`
- **Password:** `test123`

#### Demo Account
- **Email:** `demo@nodepath.com`
- **Password:** `demo123`

## Access the Application

1. Open your browser and go to: `http://localhost:8080`
2. Click on "Login" or go to: `http://localhost:8080/login`
3. Enter one of the credentials above
4. Click "Sign In"

## After Login

Once logged in, you can:
- Access the Dashboard
- View Analytics (will show sample data)
- Access Device Settings
- Use the Flow Builder
- Manage AI WhatsApp conversations

## Analytics Page

The Analytics page will now display sample data since the database is not connected. The sample data includes:
- Total Conversations
- AI Active Conversations
- Human Takeovers
- Daily breakdown charts
- Stage distribution

## Troubleshooting

### If login fails:
1. Make sure you're using the exact email and password from above (case-sensitive)
2. Clear your browser cookies and try again
3. Check that the server is running (you should see the Fiber banner in the console)

### If you need database access:
The database is currently blocking connections from your IP (124.82.240.232). To fix this:
1. Contact your database administrator to whitelist your IP
2. Or use a VPN/proxy to connect from an allowed IP
3. Or deploy the application to a server with database access

## Development Notes

The fallback authentication is only for development. In production:
- Remove the hardcoded credentials from `auth_handlers.go`
- Ensure proper database connection
- Use environment variables for any fallback credentials
