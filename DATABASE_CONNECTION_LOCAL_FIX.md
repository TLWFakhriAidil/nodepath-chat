# Database Connection Fix for Local Development

## Problem
The MySQL database at `157.245.206.124` is blocking connections from your local IP address. This prevents local development and testing.

## Solution Options

### Option 1: Use Railway Database (Recommended)
Railway provides a MySQL database that can be accessed from anywhere.

1. Go to your Railway project dashboard
2. Add a MySQL database service if you don't have one
3. Get the connection string from Railway
4. Update your `.env` file with the Railway database URL

### Option 2: Setup Local MySQL Database
1. Install MySQL locally (using Docker or direct installation)
2. Create a database named `admin_railway`
3. Update `.env` to point to local database:
```
MYSQL_URI=mysql://root:password@localhost:3306/admin_railway
DATABASE_URL=root:password@tcp(localhost:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local
```

### Option 3: Use SSH Tunnel to Remote Database
1. If you have SSH access to the server, create an SSH tunnel:
```bash
ssh -L 3306:localhost:3306 user@157.245.206.124
```
2. Then connect to `localhost:3306` instead of the remote IP

### Option 4: Request IP Whitelist
Contact the database administrator to whitelist your IP address: `124.82.240.232`

## Current Workaround for Testing

Since you want the code to be the same for localhost and Railway, here's what you can do:

1. **For Authentication**: The app has fallback authentication when database is unavailable:
   - Email: `admin@nodepath.com` Password: `admin123`
   - Email: `test@nodepath.com` Password: `test123`
   - Email: `demo@nodepath.com` Password: `demo123`

2. **For Analytics**: Without database, the analytics will show an error. To fix this properly:
   - Setup a local database (Option 2)
   - Or use Railway's database (Option 1)

## Setting Up Local MySQL with Docker (Quick Solution)

```bash
# 1. Run MySQL in Docker
docker run -d \
  --name mysql-local \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=admin_railway \
  -p 3306:3306 \
  mysql:8.0

# 2. Update .env file
MYSQL_URI=mysql://root:root@localhost:3306/admin_railway
DATABASE_URL=root:root@tcp(localhost:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local

# 3. Run migrations
go run cmd/server/main.go
```

The application will automatically create the necessary tables on startup.
