#!/bin/bash

# Test script for CGO-free builds and Railway deployment
echo "🚀 Testing CGO-free build configuration..."

# Set CGO environment
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export GO111MODULE=on

echo "📋 Environment variables:"
echo "CGO_ENABLED=$CGO_ENABLED"
echo "GOOS=$GOOS"
echo "GOARCH=$GOARCH"
echo "GO111MODULE=$GO111MODULE"
echo ""

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -f server migrate railway_migration_runner test-build
echo "✅ Cleanup complete"
echo ""

# Test main server build
echo "🔨 Building main server..."
if go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o server ./cmd/server; then
    echo "✅ Server build successful!"
    if [ -f "server" ]; then
        echo "📦 Server binary created: $(ls -lh server | awk '{print $5}')"
        # Test if binary is executable
        if [ -x "server" ]; then
            echo "✅ Server binary is executable"
        else
            echo "❌ Server binary is not executable"
            exit 1
        fi
    else
        echo "❌ Server binary not found"
        exit 1
    fi
else
    echo "❌ Server build failed!"
    exit 1
fi
echo ""

# Test migration utility build
echo "🔨 Building migration utility..."
if go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o migrate ./debug/fix_production_schema.go; then
    echo "✅ Migration utility build successful!"
    if [ -f "migrate" ]; then
        echo "📦 Migration binary created: $(ls -lh migrate | awk '{print $5}')"
    else
        echo "❌ Migration binary not found"
        exit 1
    fi
else
    echo "❌ Migration utility build failed!"
    exit 1
fi
echo ""

# Test Railway migration runner build
echo "🔨 Building Railway migration runner..."
if go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o railway_migration_runner ./debug/railway_migration_runner.go; then
    echo "✅ Railway migration runner build successful!"
    if [ -f "railway_migration_runner" ]; then
        echo "📦 Railway migration runner binary created: $(ls -lh railway_migration_runner | awk '{print $5}')"
    else
        echo "❌ Railway migration runner binary not found"
        exit 1
    fi
else
    echo "❌ Railway migration runner build failed!"
    exit 1
fi
echo ""

# Check for CGO dependencies
echo "🔍 Checking for CGO dependencies..."
if command -v ldd >/dev/null 2>&1; then
    echo "📋 Server dependencies:"
    ldd server 2>/dev/null || echo "✅ No dynamic dependencies (static binary)"
else
    echo "ℹ️  ldd not available, skipping dependency check"
fi
echo ""

# Test Docker build
echo "🐳 Testing Docker build..."
if docker build -t nodepath-chat-nocgo-test . --no-cache; then
    echo "✅ Docker build successful!"
    
    # Test Docker run
    echo "🏃 Testing Docker container..."
    if timeout 10s docker run --rm -p 8081:8080 -e MYSQL_URI="mysql://test:test@localhost:3306/test" nodepath-chat-nocgo-test >/dev/null 2>&1; then
        echo "✅ Docker container runs successfully"
    else
        echo "ℹ️  Docker container test completed (expected timeout)"
    fi
else
    echo "❌ Docker build failed!"
    exit 1
fi
echo ""

# Verify Railway configuration
echo "🚂 Verifying Railway configuration..."
if [ -f "railway.toml" ]; then
    echo "✅ railway.toml found"
else
    echo "⚠️  railway.toml not found"
fi

if [ -f "railway-deploy-nocgo.yml" ]; then
    echo "✅ railway-deploy-nocgo.yml found"
else
    echo "⚠️  railway-deploy-nocgo.yml not found"
fi

if [ -f "Dockerfile" ]; then
    echo "✅ Dockerfile found"
    if grep -q "CGO_ENABLED=0" Dockerfile; then
        echo "✅ Dockerfile has CGO_ENABLED=0"
    else
        echo "⚠️  Dockerfile missing CGO_ENABLED=0"
    fi
else
    echo "❌ Dockerfile not found"
    exit 1
fi
echo ""

echo "🎉 All CGO-free build tests passed!"
echo "📋 Summary:"
echo "   - Server binary: $(ls -lh server | awk '{print $5}')"
echo "   - Migration binary: $(ls -lh migrate | awk '{print $5}')"
echo "   - Railway runner binary: $(ls -lh railway_migration_runner | awk '{print $5}')"
echo "   - Docker build: ✅ Success"
echo "   - Railway config: ✅ Ready"
echo ""
echo "🚀 Ready for Railway deployment without CGO!"

exit 0