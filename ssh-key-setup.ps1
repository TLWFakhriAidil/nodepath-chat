# SSH Key Setup Script for Railway SSH Tunnel (Windows PowerShell)
# This script helps generate and configure SSH keys for the tunnel service

Write-Host "=== SSH Key Setup for Railway SSH Tunnel ===" -ForegroundColor Green
Write-Host "This script will help you set up SSH keys for secure database connection" -ForegroundColor Yellow
Write-Host ""

# Configuration
$KeyName = "railway-tunnel-key"
$KeyDir = "./ssh-keys"
$KeyPath = "$KeyDir/$KeyName"
$PubKeyPath = "$KeyPath.pub"

# Create SSH keys directory
Write-Host "Creating SSH keys directory..." -ForegroundColor Cyan
if (!(Test-Path $KeyDir)) {
    New-Item -ItemType Directory -Path $KeyDir -Force | Out-Null
}

# Check if key already exists
if (Test-Path $KeyPath) {
    Write-Host "SSH key already exists at $KeyPath" -ForegroundColor Yellow
    $overwrite = Read-Host "Do you want to overwrite it? (y/N)"
    if ($overwrite -ne "y" -and $overwrite -ne "Y") {
        Write-Host "Using existing key..." -ForegroundColor Green
    } else {
        Write-Host "Removing existing key..." -ForegroundColor Yellow
        Remove-Item $KeyPath -ErrorAction SilentlyContinue
        Remove-Item $PubKeyPath -ErrorAction SilentlyContinue
    }
}

# Generate SSH key if it doesn't exist
if (!(Test-Path $KeyPath)) {
    Write-Host "Generating new SSH key pair..." -ForegroundColor Cyan
    
    # Check if ssh-keygen is available
    try {
        ssh-keygen -t rsa -b 4096 -f $KeyPath -N "" -C "railway-tunnel@nodepath-chat"
        Write-Host "SSH key pair generated successfully!" -ForegroundColor Green
    } catch {
        Write-Host "Error: ssh-keygen not found. Please install OpenSSH or Git for Windows." -ForegroundColor Red
        Write-Host "You can install OpenSSH via: Add-WindowsCapability -Online -Name OpenSSH.Client~~~~0.0.1.0" -ForegroundColor Yellow
        exit 1
    }
}

# Display public key
Write-Host ""
Write-Host "=== PUBLIC KEY (copy this to your SSH server) ===" -ForegroundColor Green
Write-Host "File: $PubKeyPath" -ForegroundColor Gray
Write-Host ""
Get-Content $PubKeyPath
Write-Host ""
Write-Host "=== END PUBLIC KEY ===" -ForegroundColor Green
Write-Host ""

# Display private key for Railway
Write-Host "=== PRIVATE KEY (add this to Railway secrets as SSH_PRIVATE_KEY) ===" -ForegroundColor Green
Write-Host "File: $KeyPath" -ForegroundColor Gray
Write-Host ""
Write-Host "Copy the entire content below (including BEGIN/END lines):" -ForegroundColor Yellow
Write-Host ""
Get-Content $KeyPath
Write-Host ""
Write-Host "=== END PRIVATE KEY ===" -ForegroundColor Green
Write-Host ""

# Instructions
Write-Host "=== SETUP INSTRUCTIONS ===" -ForegroundColor Magenta
Write-Host ""
Write-Host "1. COPY PUBLIC KEY TO YOUR SSH SERVER:" -ForegroundColor Cyan
Write-Host "   Run this command on your SSH server (replace user@server):" -ForegroundColor Gray
$publicKeyContent = Get-Content $PubKeyPath
Write-Host "   echo '$publicKeyContent' >> ~/.ssh/authorized_keys" -ForegroundColor White
Write-Host ""
Write-Host "2. TEST SSH CONNECTION:" -ForegroundColor Cyan
Write-Host "   ssh -i $KeyPath user@your-server.com" -ForegroundColor White
Write-Host ""
Write-Host "3. ADD PRIVATE KEY TO RAILWAY:" -ForegroundColor Cyan
Write-Host "   - Go to Railway dashboard > Your Project > Variables" -ForegroundColor Gray
Write-Host "   - Add new variable: SSH_PRIVATE_KEY" -ForegroundColor Gray
Write-Host "   - Paste the entire private key content (including BEGIN/END lines)" -ForegroundColor Gray
Write-Host ""
Write-Host "4. UPDATE RAILWAY ENVIRONMENT VARIABLES:" -ForegroundColor Cyan
Write-Host "   TUNNEL_HOST=your-server.com" -ForegroundColor White
Write-Host "   TUNNEL_USER=your-username" -ForegroundColor White
Write-Host "   TUNNEL_PORT=22" -ForegroundColor White
Write-Host ""
Write-Host "5. TEST MYSQL CONNECTION FROM YOUR SSH SERVER:" -ForegroundColor Cyan
Write-Host "   mysql -h 159.89.198.71 -P 3306 -u admin_aqil -p admin_railway" -ForegroundColor White
Write-Host ""

# Create a test script
Write-Host "Creating test script..." -ForegroundColor Cyan
$testScript = @'
# Test script for SSH tunnel connection (PowerShell)

Write-Host "Testing SSH connection..." -ForegroundColor Cyan

$SSHKey = "./railway-tunnel-key"
$SSHUser = $env:TUNNEL_USER
if (!$SSHUser) { $SSHUser = "your-username" }
$SSHHost = $env:TUNNEL_HOST
if (!$SSHHost) { $SSHHost = "your-server.com" }
$SSHPort = $env:TUNNEL_PORT
if (!$SSHPort) { $SSHPort = "22" }

if (!(Test-Path $SSHKey)) {
    Write-Host "Error: SSH key not found at $SSHKey" -ForegroundColor Red
    Write-Host "Run ssh-key-setup.ps1 first" -ForegroundColor Yellow
    exit 1
}

Write-Host "Testing SSH connection to $SSHUser@$SSHHost:$SSHPort..." -ForegroundColor Yellow
try {
    ssh -i $SSHKey -o StrictHostKeyChecking=no -o ConnectTimeout=10 "$SSHUser@$SSHHost" "echo 'SSH connection successful!'"
    Write-Host "SSH connection test passed!" -ForegroundColor Green
} catch {
    Write-Host "SSH connection test failed: $_" -ForegroundColor Red
    exit 1
}

Write-Host "Testing MySQL connectivity from SSH server..." -ForegroundColor Yellow
try {
    ssh -i $SSHKey -o StrictHostKeyChecking=no "$SSHUser@$SSHHost" "timeout 10 bash -c '</dev/tcp/159.89.198.71/3306' && echo 'MySQL port is reachable' || echo 'MySQL port is not reachable'"
} catch {
    Write-Host "MySQL connectivity test failed: $_" -ForegroundColor Red
}

Write-Host "Testing SSH tunnel locally..." -ForegroundColor Yellow
Write-Host "Starting local tunnel (press Ctrl+C to stop)..." -ForegroundColor Cyan
try {
    ssh -i $SSHKey -o StrictHostKeyChecking=no -L 3307:159.89.198.71:3306 "$SSHUser@$SSHHost" "echo 'Tunnel established. Test with: mysql -h 127.0.0.1 -P 3307 -u admin_aqil -p admin_railway'"
} catch {
    Write-Host "SSH tunnel test failed: $_" -ForegroundColor Red
}
'@

$testScript | Out-File -FilePath "$KeyDir/test-connection.ps1" -Encoding UTF8

Write-Host "=== NEXT STEPS ===" -ForegroundColor Magenta
Write-Host ""
Write-Host "1. Copy the public key to your SSH server" -ForegroundColor White
Write-Host "2. Test the connection: cd $KeyDir && .\test-connection.ps1" -ForegroundColor White
Write-Host "3. Add the private key to Railway secrets" -ForegroundColor White
Write-Host "4. Deploy the SSH tunnel service" -ForegroundColor White
Write-Host ""
Write-Host "Files created:" -ForegroundColor Cyan
Write-Host "  - $KeyPath (private key)" -ForegroundColor Gray
Write-Host "  - $PubKeyPath (public key)" -ForegroundColor Gray
Write-Host "  - $KeyDir/test-connection.ps1 (test script)" -ForegroundColor Gray
Write-Host ""
Write-Host "Setup complete! 🚀" -ForegroundColor Green

# Copy private key to clipboard if possible
try {
    Get-Content $KeyPath | Set-Clipboard
    Write-Host "Private key copied to clipboard! You can paste it directly into Railway." -ForegroundColor Yellow
} catch {
    Write-Host "Note: Could not copy to clipboard. Please copy the private key manually." -ForegroundColor Yellow
}