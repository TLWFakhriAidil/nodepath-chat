# SSH Tunnel Setup for Secure MySQL Connection on Railway

This guide explains how to set up a secure SSH tunnel to connect to the external MySQL database at `159.89.198.71:3306` without using static IPs.

## Overview

Instead of using Railway's static IP feature (which requires a Pro plan), we use an SSH tunnel through a jump server to securely connect to the external MySQL database. This approach maintains security while avoiding the need for IP whitelisting.

## Architecture

```
Railway App → SSH Tunnel Container → Jump Server → MySQL Database
                                   (your server)   (159.89.198.71:3306)
```

## Prerequisites

1. **SSH Jump Server**: You need access to an SSH server that can reach the MySQL database
2. **SSH Key Pair**: Generate SSH keys for secure authentication
3. **Railway Account**: For deploying the tunnel service

## Step 1: Prepare SSH Jump Server

### Option A: Use Your Own Server
If you have a VPS or server that can reach `159.89.198.71:3306`:

```bash
# Test connectivity from your server
telnet 159.89.198.71 3306
# or
nc -zv 159.89.198.71 3306
```

### Option B: Use Cloud Provider
Deploy a small VPS on:
- **DigitalOcean**: $5/month droplet
- **AWS EC2**: t3.nano instance
- **Google Cloud**: e2-micro instance
- **Linode**: Nanode 1GB

## Step 2: Generate SSH Keys

```bash
# Generate SSH key pair
ssh-keygen -t rsa -b 4096 -f ssh-tunnel-key -N ""

# This creates:
# - ssh-tunnel-key (private key)
# - ssh-tunnel-key.pub (public key)
```

## Step 3: Configure Jump Server

1. **Copy public key to jump server**:
```bash
ssh-copy-id -i ssh-tunnel-key.pub user@your-jump-server.com
```

2. **Test SSH connection**:
```bash
ssh -i ssh-tunnel-key user@your-jump-server.com
```

3. **Test MySQL connectivity from jump server**:
```bash
# From your jump server
mysql -h 159.89.198.71 -P 3306 -u admin_aqil -p admin_railway
```

## Step 4: Deploy SSH Tunnel on Railway

### 4.1 Create New Railway Service

1. Go to Railway dashboard
2. Create new project or add service to existing project
3. Choose "Deploy from GitHub repo" or "Empty Service"

### 4.2 Configure Environment Variables

Add these environment variables in Railway:

```env
# SSH server details (replace with your values)
TUNNEL_HOST=your-jump-server.com
TUNNEL_PORT=22
TUNNEL_USER=your-ssh-username

# Tunnel configuration
LOCAL_PORT=3307
REMOTE_HOST=159.89.198.71
REMOTE_PORT=3306

# SSH private key (paste the entire private key content)
SSH_PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
(paste your private key content here)
-----END RSA PRIVATE KEY-----
```

### 4.3 Deploy the Tunnel Service

1. Upload the `Dockerfile.tunnel` to your repository
2. Set Railway to use `Dockerfile.tunnel` for build
3. Deploy the service

## Step 5: Update Main Application

### 5.1 Update Database Connection

Modify your main application's environment variables:

```env
# Use the tunnel service for database connection
DATABASE_URL=mysql://admin_aqil:admin_aqil@ssh-tunnel.railway.internal:3307/admin_railway?charset=utf8mb4&parseTime=True&loc=Local
MYSQL_URI=mysql://admin_aqil:admin_aqil@ssh-tunnel.railway.internal:3307/admin_railway?charset=utf8mb4&parseTime=True&loc=Local
```

### 5.2 Update Railway Service Dependencies

In your main application's Railway settings:
1. Add the SSH tunnel service as a dependency
2. Ensure the tunnel service starts before your main app

## Step 6: Testing

### 6.1 Test Tunnel Service

1. Check Railway logs for the tunnel service
2. Look for "SSH tunnel connected" messages
3. Verify health checks are passing

### 6.2 Test Database Connection

1. Deploy your main application
2. Check application logs for database connection
3. Test database operations through your app

## Security Benefits

✅ **No Static IP Required**: Avoids Railway Pro plan requirement
✅ **Encrypted Connection**: All traffic goes through SSH tunnel
✅ **No Direct Exposure**: Database not directly accessible from internet
✅ **Authentication**: Uses SSH key authentication
✅ **Monitoring**: Can monitor tunnel health and connection status

## Troubleshooting

### Common Issues

1. **SSH Connection Failed**
   - Verify SSH key is correct
   - Check jump server accessibility
   - Ensure SSH user has proper permissions

2. **Tunnel Port Not Accessible**
   - Check Railway service logs
   - Verify environment variables
   - Test jump server to MySQL connectivity

3. **Application Can't Connect**
   - Verify service dependencies in Railway
   - Check internal Railway networking
   - Ensure correct database URL format

### Debug Commands

```bash
# Test SSH connection manually
ssh -i ssh-tunnel-key -L 3307:159.89.198.71:3306 user@jump-server.com

# Test local tunnel
mysql -h 127.0.0.1 -P 3307 -u admin_aqil -p admin_railway

# Check port accessibility
nc -zv localhost 3307
```

## Cost Comparison

| Solution | Monthly Cost | Security | Complexity |
|----------|-------------|----------|------------|
| Railway Pro + Static IP | $20+ | Medium | Low |
| SSH Tunnel + Small VPS | $5-10 | High | Medium |
| This SSH Tunnel Solution | $5-10 | High | Medium |

## Alternative Jump Server Options

If you don't have a jump server, consider:

1. **Free Tier Options**:
   - AWS EC2 Free Tier (12 months)
   - Google Cloud Free Tier
   - Oracle Cloud Always Free

2. **Low-Cost Options**:
   - DigitalOcean $5/month droplet
   - Linode $5/month nanode
   - Vultr $2.50/month instance

## Next Steps

1. Set up your SSH jump server
2. Generate and configure SSH keys
3. Deploy the tunnel service on Railway
4. Update your main application configuration
5. Test the complete setup
6. Monitor tunnel health and performance

This solution provides a secure, cost-effective way to connect to your external MySQL database without requiring static IPs or exposing the database directly to the internet.