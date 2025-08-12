# Local & Railway Environment Synchronization

This guide ensures your local development environment matches Railway's production deployment exactly.

## 🎯 Problem

Previously, there was a mismatch between:
- **Local Development**: Running `npm run dev` (Vite dev server) + `go run cmd/server/main.go`
- **Railway Production**: Built React app served by Go backend from `dist/` folder

## ✅ Solution

Now your local environment can match Railway's deployment exactly.

## 🚀 Quick Start

### Option 1: Use NPM Scripts (Recommended)
```bash
# Build frontend and start Go server (matches Railway)
npm run railway:local

# Alternative command
npm run serve:production
```

### Option 2: Use Shell Scripts

**For Windows (PowerShell):**
```powershell
.\build-and-serve.ps1
```

**For Linux/Mac (Bash):**
```bash
./build-and-serve.sh
```

### Option 3: Manual Steps
```bash
# 1. Build the frontend
npm run build

# 2. Start the Go backend (serves from dist/)
go run cmd/server/main.go
```

## 🔍 How It Works

### Railway Deployment Process
1. **Frontend Build**: `npm run build` → Creates `dist/` folder
2. **Backend Compilation**: `go build` → Creates Go binary
3. **Static Serving**: Go server serves React app from `dist/`
4. **API Routes**: Go server handles `/api/*` routes
5. **SPA Routing**: Catch-all route serves `dist/index.html`

### Local Environment (Now Matching)
1. **Frontend Build**: `npm run build` → Creates `dist/` folder
2. **Backend Running**: `go run cmd/server/main.go`
3. **Static Serving**: Go server serves React app from `dist/`
4. **API Routes**: Go server handles `/api/*` routes
5. **SPA Routing**: Catch-all route serves `dist/index.html`

## 📁 File Structure

```
├── dist/                    # Built React app (served by Go)
│   ├── index.html
│   ├── assets/
│   └── ...
├── src/                     # React source code
├── cmd/server/main.go       # Go backend entry point
├── build-and-serve.sh       # Linux/Mac script
├── build-and-serve.ps1      # Windows PowerShell script
└── package.json             # Updated with new scripts
```

## 🔧 Configuration Details

### Go Server Configuration
- **Static Files**: `app.Static("/", "./dist")`
- **API Routes**: `app.Group("/api")`
- **SPA Fallback**: `app.Get("/*", serve index.html)`
- **Port**: 8080 (matches Railway)

### Build Process
- **Vite Build**: Optimized production build
- **Output**: `dist/` folder
- **Assets**: Hashed filenames for caching

## 🎯 Benefits

1. **Exact Match**: Local environment identical to Railway
2. **Production Testing**: Test production build locally
3. **Performance**: Same optimizations as production
4. **Debugging**: Catch deployment issues early
5. **Consistency**: No environment-specific bugs

## 🔄 Development Workflow

### For Active Development
```bash
# Use development server for hot reload
npm run dev
```

### For Production Testing
```bash
# Use production build (matches Railway)
npm run railway:local
```

### Before Deployment
```bash
# Always test with production build
npm run railway:local
# Verify everything works, then deploy to Railway
```

## 🚨 Important Notes

1. **Redis Warning**: Local Redis connection may fail (normal for development)
2. **Environment Variables**: Ensure `.env` file matches Railway settings
3. **Database**: Use same database configuration as Railway
4. **Port**: Always uses 8080 (Railway's default)

## 🔍 Troubleshooting

### Build Fails
```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
npm run build
```

### Go Server Issues
```bash
# Check Go modules
go mod tidy
go run cmd/server/main.go
```

### Static Files Not Loading
- Ensure `dist/` folder exists after build
- Check Go server logs for file serving errors
- Verify paths in `cmd/server/main.go`

Now your local environment perfectly matches Railway's production deployment! 🎉