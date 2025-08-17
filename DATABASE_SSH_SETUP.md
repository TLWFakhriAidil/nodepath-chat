# Database Server SSH Setup Guide

This guide explains how to configure SSH access on the database server (159.89.198.71) to allow secure tunnel connections from Railway.

## Overview

The SSH tunnel solution eliminates the need for IP whitelisting by creating a secure encrypted connection between Railway and the database server. This approach is:

- ✅ **Scalable**: Handles 3000+ concurrent connections
- ✅ **Secure**: Encrypted SSH tunnel with key authentication
- ✅ **Reliable**: Not affected by Railway's dynamic IP changes
- ✅ **Cost-effective**: No need for Railway Pro plan
- ✅ **Future-proof**: Works regardless of Railway infrastructure changes

## Prerequisites

1. **Root/sudo access** to database server (159.89.198.71)
2. **SSH access** to the database server
3. **MySQL running** on the database server
4. **Railway project** with SSH tunnel service configured

## Step 1: Generate SSH Keys

Run the SSH key generation script:

### Windows (PowerShell)
```powershell
.\ssh-key-setup.ps1
```

### Linux/Mac (Bash)
```bash
./ssh-key-setup.sh
```

This will create:
- `ssh-keys/railway-tunnel-key` (private key for Railway)
- `ssh-keys/railway-tunnel-key.pub` (public key for database server)

## Step 2: Configure Database Server SSH Access

### 2.1 Connect to Database Server
```bash
ssh root@159.89.198.71
```

### 2.2 Create Tunnel User (Recommended)
```bash
# Create dedicated user for tunnel connections
useradd -m -s /bin/bash railway-tunnel

# Create SSH directory
sudo -u railway-tunnel mkdir -p /home/railway-tunnel/.ssh
sudo -u railway-tunnel chmod 700 /home/railway-tunnel/.ssh
```

### 2.3 Add Public Key
```bash
# Copy the public key content from ssh-keys/railway-tunnel-key.pub
# and add it to authorized_keys
sudo -u railway-tunnel nano /home/railway-tunnel/.ssh/authorized_keys

# Paste the public key content, then save and exit

# Set proper permissions
sudo -u railway-tunnel chmod 600 /home/railway-tunnel/.ssh/authorized_keys
```

### 2.4 Configure SSH for Tunnel User
```bash
# Edit SSH configuration
sudo nano /etc/ssh/sshd_config

# Add these lines at the end:
Match User railway-tunnel
    AllowTcpForwarding yes
    X11Forwarding no
    PermitTunnel no
    GatewayPorts no
    AllowAgentForwarding no
    PermitOpen localhost:3306
    ForceCommand echo 'This account can only be used for SSH tunneling'
```

### 2.5 Restart SSH Service
```bash
sudo systemctl restart sshd
```

## Step 3: Test SSH Connection

### 3.1 Test from Local Machine
```bash
# Test SSH connection
ssh -i ssh-keys/railway-tunnel-key railway-tunnel@159.89.198.71

# Test SSH tunnel
ssh -i ssh-keys/railway-tunnel-key -L 3307:localhost:3306 railway-tunnel@159.89.198.71
```

### 3.2 Test MySQL Connection Through Tunnel
```bash
# In another terminal, test MySQL connection
mysql -h 127.0.0.1 -P 3307 -u admin_aqil -p admin_railway
```

## Step 4: Configure Railway Environment

### 4.1 Add SSH Private Key to Railway
1. Go to Railway dashboard → Your Project → Variables
2. Add new variable: `SSH_PRIVATE_KEY`
3. Paste the entire private key content (including BEGIN/END lines)

### 4.2 Set Tunnel Environment Variables
```bash
TUNNEL_HOST=159.89.198.71
TUNNEL_USER=railway-tunnel
TUNNEL_PORT=22
LOCAL_PORT=3306
REMOTE_HOST=127.0.0.1
REMOTE_PORT=3306
```

## Step 5: Update Database Access Control

### 5.1 Remove IP-based Restrictions
```sql
-- Remove previous IP-based grants
REVOKE ALL PRIVILEGES ON admin_railway.* FROM 'admin_aqil'@'113.211.0.0/255.255.0.0';
REVOKE ALL PRIVILEGES ON admin_railway.* FROM 'admin_aqil'@'208.77.0.0/255.255.0.0';
REVOKE ALL PRIVILEGES ON admin_railway.* FROM 'admin_aqil'@'113.211.%';
REVOKE ALL PRIVILEGES ON admin_railway.* FROM 'admin_aqil'@'208.77.%';
REVOKE ALL PRIVILEGES ON admin_railway.* FROM 'admin_aqil'@'%.railway.app';

-- Clean up user entries
DROP USER IF EXISTS 'admin_aqil'@'113.211.0.0/255.255.0.0';
DROP USER IF EXISTS 'admin_aqil'@'208.77.0.0/255.255.0.0';
DROP USER IF EXISTS 'admin_aqil'@'113.211.%';
DROP USER IF EXISTS 'admin_aqil'@'208.77.%';
DROP USER IF EXISTS 'admin_aqil'@'%.railway.app';

FLUSH PRIVILEGES;
```

### 5.2 Grant Localhost Access Only
```sql
-- Grant access only from localhost (tunnel connections)
GRANT ALL PRIVILEGES ON admin_railway.* TO 'admin_aqil'@'localhost';
GRANT ALL PRIVILEGES ON admin_railway.* TO 'admin_aqil'@'127.0.0.1';
FLUSH PRIVILEGES;
```

## Step 6: Deploy and Test

### 6.1 Deploy SSH Tunnel Service
1. Ensure `railway-deploy.yml` has SSH tunnel service enabled
2. Deploy to Railway
3. Monitor logs for successful tunnel establishment

### 6.2 Deploy Main Application
1. Update database connection to use tunnel
2. Deploy main application
3. Test database connectivity

## Security Benefits

✅ **Encrypted Connection**: All database traffic encrypted via SSH
✅ **Key-based Authentication**: No password-based access
✅ **Restricted User**: Tunnel user can only create tunnels
✅ **Localhost Only**: Database only accepts local connections
✅ **No IP Management**: Eliminates dynamic IP whitelisting issues

## Troubleshooting

### SSH Connection Issues
```bash
# Test SSH connection with verbose output
ssh -v -i ssh-keys/railway-tunnel-key railway-tunnel@159.89.198.71

# Check SSH server logs
sudo tail -f /var/log/auth.log
```

### Tunnel Issues
```bash
# Test tunnel manually
ssh -v -i ssh-keys/railway-tunnel-key -L 3307:localhost:3306 railway-tunnel@159.89.198.71

# Check if port is listening
netstat -tlnp | grep 3307
```

### Database Connection Issues
```bash
# Test MySQL connectivity
mysql -h 127.0.0.1 -P 3307 -u admin_aqil -p admin_railway

# Check MySQL logs
sudo tail -f /var/log/mysql/error.log
```

## Monitoring

### Railway Logs
- Monitor SSH tunnel service logs
- Look for "SSH tunnel connected" messages
- Check for connection retry attempts

### Database Server
```bash
# Monitor SSH connections
sudo netstat -tnp | grep :22

# Monitor MySQL connections
sudo netstat -tnp | grep :3306

# Check active SSH tunnels
sudo ss -tlnp | grep :3306
```

## Maintenance

### Key Rotation
1. Generate new SSH key pair
2. Add new public key to authorized_keys
3. Update Railway SSH_PRIVATE_KEY variable
4. Remove old public key after verification

### Monitoring Health
- Set up alerts for SSH tunnel disconnections
- Monitor database connection metrics
- Regular testing of tunnel connectivity

This SSH tunnel approach provides a robust, secure, and scalable solution for connecting Railway to your external MySQL database without the limitations of IP-based whitelisting.