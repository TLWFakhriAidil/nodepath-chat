# Redis Local Setup Script for NodePath Chat System
# This script helps set up Redis for local development

Write-Host "Redis Local Setup for NodePath Chat" -ForegroundColor Green
Write-Host "====================================" -ForegroundColor Green

# Function to check if Docker is installed
function Test-DockerInstalled {
    try {
        docker --version | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

# Function to start Redis with Docker
function Start-RedisDocker {
    Write-Host "Starting Redis with Docker..." -ForegroundColor Yellow
    
    # Create docker-compose.yml for Redis
    $dockerCompose = @"
version: '3.8'
services:
  redis:
    image: redis:7-alpine
    container_name: nodepath-redis
    ports:
      - "6379:6379"
    command: redis-server --requirepass nodepath123
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  redis_data:
"@
    
    $dockerCompose | Out-File -FilePath "docker-compose-redis.yml" -Encoding UTF8
    
    # Start Redis container
    docker-compose -f docker-compose-redis.yml up -d
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Redis started successfully with Docker!" -ForegroundColor Green
        Write-Host "Connection Details:" -ForegroundColor Cyan
        Write-Host "   URL: redis://default:nodepath123@localhost:6379" -ForegroundColor White
        Write-Host "   Password: nodepath123" -ForegroundColor White
        Write-Host "   Port: 6379" -ForegroundColor White
        return $true
    } else {
        Write-Host "Failed to start Redis with Docker" -ForegroundColor Red
        return $false
    }
}

# Function to test Redis connection
function Test-RedisConnection {
    Write-Host "Testing Redis connection..." -ForegroundColor Yellow
    
    try {
        $result = docker exec nodepath-redis redis-cli -a nodepath123 ping 2>$null
        if ($result -eq "PONG") {
            Write-Host "Redis connection successful!" -ForegroundColor Green
            return $true
        }
        
        Write-Host "Redis connection failed" -ForegroundColor Red
        return $false
    }
    catch {
        Write-Host "Error testing Redis connection" -ForegroundColor Red
        return $false
    }
}

# Main execution
try {
    # Check if Docker is available
    if (Test-DockerInstalled) {
        Write-Host "Docker found" -ForegroundColor Green
        
        # Start Redis with Docker
        if (Start-RedisDocker) {
            Start-Sleep -Seconds 3
            
            # Test connection
            if (Test-RedisConnection) {
                Write-Host "" -ForegroundColor White
                Write-Host "Redis setup completed successfully!" -ForegroundColor Green
                Write-Host "" -ForegroundColor White
                Write-Host "Environment variables are already set in .env file" -ForegroundColor Cyan
                Write-Host "" -ForegroundColor White
                Write-Host "To start the application:" -ForegroundColor Cyan
                Write-Host "   go run cmd/server/main.go" -ForegroundColor White
                Write-Host "" -ForegroundColor White
                Write-Host "Redis is running in the background" -ForegroundColor Green
                Write-Host "To stop Redis: docker-compose -f docker-compose-redis.yml down" -ForegroundColor Yellow
            } else {
                Write-Host "Redis connection test failed" -ForegroundColor Red
                Write-Host "Try running: docker-compose -f docker-compose-redis.yml logs redis" -ForegroundColor Yellow
            }
        }
    } else {
        Write-Host "Docker not found" -ForegroundColor Red
        Write-Host "" -ForegroundColor White
        Write-Host "Please install Docker Desktop:" -ForegroundColor Yellow
        Write-Host "   https://www.docker.com/products/docker-desktop" -ForegroundColor White
        Write-Host "" -ForegroundColor White
        Write-Host "Alternative: Install Redis directly:" -ForegroundColor Yellow
        Write-Host "   https://redis.io/download" -ForegroundColor White
        Write-Host "" -ForegroundColor White
        Write-Host "Environment variables are set in .env file for manual Redis setup" -ForegroundColor Yellow
    }
}
catch {
    Write-Host "Error during setup: $_" -ForegroundColor Red
    Write-Host "" -ForegroundColor White
    Write-Host "Manual setup instructions:" -ForegroundColor Yellow
    Write-Host "1. Install Docker Desktop" -ForegroundColor White
    Write-Host "2. Run: docker run -d -p 6379:6379 --name nodepath-redis redis:7-alpine redis-server --requirepass nodepath123" -ForegroundColor White
    Write-Host "3. Update REDIS_URL in .env file" -ForegroundColor White
    Write-Host "4. Run: go run cmd/server/main.go" -ForegroundColor White
}

Write-Host "" -ForegroundColor White
Write-Host "For more information, see REDIS_SETUP_GUIDE.md" -ForegroundColor Cyan