# CGO-Free Deployment Guide for Railway

This guide explains how to deploy the NodePath Chat application on Railway without CGO dependencies.

## Overview

The application has been configured to build and run without CGO (C bindings) to ensure compatibility with Railway's deployment environment and avoid common deployment issues.

## Configuration Changes

### 1. Dockerfile Optimizations

- **Removed build-base dependency**: No longer requires C compiler tools
- **Added CGO environment variables**: `CGO_ENABLED=0`, `GOOS=linux`, `GOARCH=amd64`
- **Static linking flags**: Added `-ldflags '-extldflags "-static"'` for fully static binaries
- **Minimal runtime dependencies**: Only essential packages (ca-certificates, tzdata, bash)

### 2. Build Configuration

```dockerfile
# Backend build stage
FROM golang:1.24-alpine AS backend-builder

# Install minimal dependencies for CGO-free builds
RUN apk add --no-cache git ca-certificates tzdata

# Set CGO environment variables for static builds
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on

# Build commands with static linking
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /app/bin/server ./cmd/server
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /app/bin/migrate ./debug/fix_production_schema.go
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /app/bin/railway_migration_runner ./debug/railway_migration_runner.go
```

### 3. Railway Configuration

Created `railway-deploy-nocgo.yml` with optimized settings:

- **Build environment**: CGO_ENABLED=0
- **Resource limits**: 1GB memory, 1 CPU core
- **Auto-scaling**: 2-10 replicas based on load
- **Health checks**: `/health` endpoint monitoring
- **Performance optimizations**: HTTP/2, compression, connection pooling

## Testing

### Local Testing Scripts

1. **Bash script**: `test-nocgo.sh` - Comprehensive Linux/Docker testing
2. **PowerShell script**: `test-nocgo.ps1` - Windows testing

Both scripts test:
- CGO-free compilation
- Binary creation and execution
- Docker build process
- Configuration file validation

### Running Tests

```bash
# Linux/macOS/WSL
bash test-nocgo.sh

# Windows PowerShell
powershell -ExecutionPolicy Bypass -File test-nocgo.ps1
```

## Deployment Steps

### 1. Verify Configuration

```bash
# Check CGO is disabled
go env CGO_ENABLED  # Should return 0

# Test local build
CGO_ENABLED=0 go build -o test-server ./cmd/server
```

### 2. Deploy to Railway

1. **Push to repository**: Ensure all changes are committed
2. **Railway deployment**: Uses Dockerfile automatically
3. **Environment variables**: Set `MYSQL_URI` and other required vars
4. **Health check**: Monitor `/health` endpoint

### 3. Verify Deployment

- Check Railway logs for successful startup
- Test health endpoint: `https://your-app.railway.app/health`
- Verify database connectivity
- Test WhatsApp webhook endpoints

## Benefits of CGO-Free Deployment

1. **Faster builds**: No C compiler dependencies
2. **Smaller images**: Reduced base image requirements
3. **Better compatibility**: Works across different Linux distributions
4. **Improved security**: Fewer system dependencies
5. **Easier debugging**: Static binaries with no external dependencies

## Performance Considerations

### High Load Handling (3000+ devices)

- **Connection pooling**: Configured in `railway-deploy-nocgo.yml`
- **Auto-scaling**: 2-10 replicas based on CPU/memory usage
- **Database optimization**: Connection limits and timeouts
- **Caching**: 5-minute TTL for frequently accessed data

### Monitoring

- **Health checks**: Every 30 seconds
- **Resource monitoring**: CPU, memory, network, database connections
- **Alerts**: Configured for 80% CPU, 85% memory, 2s response time

## Troubleshooting

### Common Issues

1. **Build failures**: Check CGO_ENABLED=0 in environment
2. **Runtime errors**: Verify static linking flags
3. **Database connection**: Ensure MYSQL_URI is correctly set
4. **Memory issues**: Monitor resource usage and adjust limits

### Debug Commands

```bash
# Check binary dependencies
ldd ./server  # Should show "not a dynamic executable" or minimal deps

# Test CGO status
go env CGO_ENABLED

# Verify build flags
go build -x ./cmd/server  # Shows detailed build process
```

## Files Modified/Created

- `Dockerfile` - Updated for CGO-free builds
- `railway-deploy-nocgo.yml` - Railway deployment configuration
- `test-nocgo.sh` - Linux/Docker testing script
- `test-nocgo.ps1` - Windows testing script
- `DEPLOYMENT-NOCGO.md` - This deployment guide

## Next Steps

1. Test deployment on Railway staging environment
2. Monitor performance with real traffic
3. Adjust auto-scaling parameters based on usage patterns
4. Implement additional monitoring and alerting

---

**Note**: This configuration ensures the application runs efficiently on Railway without CGO dependencies, providing better compatibility and performance for handling 3000+ concurrent devices.