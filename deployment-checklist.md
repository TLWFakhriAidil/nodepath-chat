# SSH Tunnel Deployment Checklist

This checklist ensures proper deployment and testing of the SSH tunnel solution for secure MySQL database connection.

## Pre-Deployment Checklist

### ✅ 1. SSH Jump Server Setup

- [ ] **SSH Server Access**: Verify you have SSH access to a server that can reach `159.89.198.71:3306`
- [ ] **MySQL Connectivity**: Test MySQL connection from your SSH server
  ```bash
  # Test from your SSH server
  telnet 159.89.198.71 3306
  # or
  mysql -h 159.89.198.71 -P 3306 -u admin_aqil -p admin_railway
  ```
- [ ] **SSH Server Requirements**: Ensure SSH server allows port forwarding

### ✅ 2. SSH Key Generation

- [ ] **Generate SSH Keys**: Run the key generation script
  ```powershell
  # Windows
  .\ssh-key-setup.ps1
  
  # Linux/Mac
  ./ssh-key-setup.sh
  ```
- [ ] **Copy Public Key**: Add public key to SSH server's `~/.ssh/authorized_keys`
- [ ] **Test SSH Connection**: Verify SSH key authentication works
  ```bash
  ssh -i ./ssh-keys/railway-tunnel-key user@your-server.com
  ```

### ✅ 3. Railway Configuration

- [ ] **Environment Variables**: Set up Railway environment variables
  ```env
  # SSH Tunnel Service
  TUNNEL_HOST=your-jump-server.com
  TUNNEL_USER=your-ssh-username
  TUNNEL_PORT=22
  LOCAL_PORT=3307
  REMOTE_HOST=159.89.198.71
  REMOTE_PORT=3306
  SSH_PRIVATE_KEY=<paste-entire-private-key>
  
  # Main Application
  DATABASE_URL=mysql://admin_aqil:admin_aqil@mysql-ssh-tunnel.railway.internal:3307/admin_railway?charset=utf8mb4&parseTime=True&loc=Local
  OPENROUTER_DEFAULT_KEY=<your-api-key>
  ```

## Deployment Steps

### 🚀 Step 1: Deploy SSH Tunnel Service

1. **Create New Railway Service**
   - [ ] Go to Railway dashboard
   - [ ] Create new service or add to existing project
   - [ ] Name it `mysql-ssh-tunnel`

2. **Configure Build Settings**
   - [ ] Set Dockerfile path to `Dockerfile.tunnel`
   - [ ] Ensure build context is project root

3. **Set Environment Variables**
   - [ ] Add all tunnel-related environment variables
   - [ ] Paste SSH private key content (including BEGIN/END lines)

4. **Deploy Service**
   - [ ] Trigger deployment
   - [ ] Monitor logs for successful tunnel establishment
   - [ ] Look for "SSH tunnel connected" messages

### 🚀 Step 2: Update Main Application

1. **Update Database Configuration**
   - [ ] Change `DATABASE_URL` to use tunnel service
   - [ ] Update `MYSQL_URI` if used
   - [ ] Ensure connection string points to `mysql-ssh-tunnel.railway.internal:3307`

2. **Set Service Dependencies**
   - [ ] Configure main app to depend on tunnel service
   - [ ] Ensure tunnel starts before main application

3. **Deploy Main Application**
   - [ ] Redeploy main application with new database configuration
   - [ ] Monitor startup logs for database connection success

## Testing & Verification

### 🧪 Test 1: SSH Tunnel Service Health

- [ ] **Check Service Status**: Verify tunnel service is running
- [ ] **Review Logs**: Look for successful SSH connection messages
  ```
  Expected log messages:
  ✅ "SSH tunnel connected"
  ✅ "Tunnel port 3307 is accessible"
  ✅ Health check passing
  ```
- [ ] **Health Check**: Verify health endpoint responds

### 🧪 Test 2: Database Connectivity

- [ ] **Application Startup**: Check main app connects to database
- [ ] **Database Operations**: Test basic CRUD operations
- [ ] **Connection Pool**: Verify connection pooling works
  ```
  Expected log messages:
  ✅ "Database connection established successfully"
  ✅ "Using MySQL for application data"
  ✅ No connection timeout errors
  ```

### 🧪 Test 3: Application Functionality

- [ ] **API Endpoints**: Test key API endpoints
- [ ] **Database Queries**: Verify data retrieval works
- [ ] **Real-time Features**: Test WebSocket connections
- [ ] **AI Integration**: Test OpenRouter API calls

### 🧪 Test 4: Performance & Reliability

- [ ] **Load Testing**: Test with multiple concurrent connections
- [ ] **Connection Recovery**: Test tunnel reconnection after failure
- [ ] **Error Handling**: Verify graceful error handling
- [ ] **Monitoring**: Set up alerts for tunnel health

## Troubleshooting Guide

### 🔧 Common Issues

#### SSH Connection Failed
```
Symptoms: "SSH connection failed" in tunnel service logs
Solutions:
- [ ] Verify SSH server is accessible
- [ ] Check SSH key format (include BEGIN/END lines)
- [ ] Ensure SSH user has proper permissions
- [ ] Test SSH connection manually
```

#### Tunnel Port Not Accessible
```
Symptoms: "Tunnel port 3307 is not accessible"
Solutions:
- [ ] Check SSH server allows port forwarding
- [ ] Verify MySQL server is reachable from SSH server
- [ ] Check firewall rules on SSH server
- [ ] Test manual tunnel: ssh -L 3307:159.89.198.71:3306 user@server
```

#### Application Can't Connect to Database
```
Symptoms: "Failed to connect to database" in app logs
Solutions:
- [ ] Verify tunnel service is running and healthy
- [ ] Check DATABASE_URL format
- [ ] Ensure service dependencies are configured
- [ ] Test internal Railway networking
```

#### Connection Timeouts
```
Symptoms: Database connection timeouts
Solutions:
- [ ] Increase connection timeout values
- [ ] Check SSH server stability
- [ ] Verify MySQL server performance
- [ ] Review connection pool settings
```

### 🔍 Debug Commands

```bash
# Test SSH connection manually
ssh -i ssh-keys/railway-tunnel-key user@your-server.com

# Test manual tunnel
ssh -i ssh-keys/railway-tunnel-key -L 3307:159.89.198.71:3306 user@your-server.com

# Test MySQL through tunnel
mysql -h 127.0.0.1 -P 3307 -u admin_aqil -p admin_railway

# Check port accessibility
nc -zv localhost 3307

# Test from SSH server
ssh user@your-server.com "telnet 159.89.198.71 3306"
```

## Monitoring & Maintenance

### 📊 Health Monitoring

- [ ] **Set up Railway Alerts**: Configure alerts for service health
- [ ] **Monitor Logs**: Regular review of tunnel and app logs
- [ ] **Performance Metrics**: Track connection times and success rates
- [ ] **Uptime Monitoring**: External monitoring of application availability

### 🔄 Maintenance Tasks

- [ ] **SSH Key Rotation**: Plan for periodic SSH key updates
- [ ] **Server Updates**: Keep SSH jump server updated
- [ ] **Backup Strategy**: Ensure SSH keys are backed up securely
- [ ] **Documentation**: Keep deployment docs updated

## Security Considerations

### 🔒 Security Checklist

- [ ] **SSH Key Security**: Private keys stored securely in Railway secrets
- [ ] **Server Hardening**: SSH server properly configured and hardened
- [ ] **Network Security**: Minimal required ports open
- [ ] **Access Control**: Limited SSH user permissions
- [ ] **Audit Logging**: SSH access logging enabled

### 🛡️ Best Practices

- [ ] **Key Rotation**: Regular SSH key rotation schedule
- [ ] **Monitoring**: Continuous monitoring of SSH connections
- [ ] **Backup Access**: Alternative access methods documented
- [ ] **Incident Response**: Plan for tunnel service failures

## Success Criteria

✅ **Deployment Successful When**:
- SSH tunnel service is running and healthy
- Main application connects to database through tunnel
- All application features work correctly
- Performance meets requirements (3000+ concurrent users)
- Monitoring and alerts are configured
- Security best practices are implemented

## Cost Analysis

| Component | Monthly Cost | Notes |
|-----------|-------------|-------|
| Railway SSH Tunnel Service | $5-10 | Small container |
| SSH Jump Server (VPS) | $5-10 | DigitalOcean/Linode |
| **Total** | **$10-20** | vs $20+ for Railway Pro |

**Benefits**:
- 50% cost savings vs Railway Pro plan
- Enhanced security with SSH encryption
- Full control over tunnel configuration
- No vendor lock-in for static IP feature

---

**Next Steps**: After successful deployment, consider implementing automated health checks and failover mechanisms for production reliability.