# MySQL Database Connection Guide

## Database Credentials

```
Host: 157.245.206.124
Port: 3306
Database: admin_railway
Username: admin_aqil
Password: admin_aqil
```

## Current System Integration

The NodePath Chat system is already configured to use these database credentials:

- **Configuration**: Located in `internal/config/config.go`
- **Environment Variables**: Set in `.env` file
- **Database Service**: Implemented in `internal/database/database.go`
- **Connection Status**: ✅ **VERIFIED WORKING**

### Database Tables

The system currently uses the following tables:

#### Core Application Tables
- `chatbot_flows_nodepath` - Stores chatbot flow configurations

- `chatbot_leads` - Lead management data
- `leads` - Additional lead information
- `leads_ai` - AI-processed lead data

#### WhatsApp Integration Tables
- `whatsapp_chats` - Chat conversations
- `whatsapp_messages` - Individual messages
- `whatsmeow_*` - WhatsApp protocol implementation tables

#### Campaign & Analytics Tables
- `campaigns` - Marketing campaigns
- `ai_campaign_progress` - AI campaign tracking
- `broadcast_messages*` - Broadcast message management
- `message_analytics` - Message performance metrics
- `sequence_*` - Automated sequence management

#### User & Device Management
- `users` - User accounts
- `team_members` - Team management
- `user_devices` - Device registrations
- `user_sessions` - Session management
- `device_load_balance` - Load balancing configuration

#### Media & Files
- `media_files` - File storage references

## Connection Methods

### 1. Using MySQL Command Line
```bash
mysql -h 157.245.206.124 -P 3306 -u admin_aqil -padmin_aqil admin_railway
```

### 2. Using phpMyAdmin
```
URL: http://157.245.206.124/phpmyadmin
Username: admin_aqil
Password: admin_aqil
```

### 3. Using Python (pymysql)
```python
import pymysql

connection = pymysql.connect(
    host='157.245.206.124',
    port=3306,
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway',
    cursorclass=pymysql.cursors.DictCursor
)

cursor = connection.cursor()
cursor.execute("SELECT * FROM broadcast_messages LIMIT 5")
results = cursor.fetchall()
for row in results:
    print(row)

cursor.close()
connection.close()
```

### 4. Using MySQL Workbench
1. Click "+" to create new connection
2. Connection Name: WhatsApp Multi-Device
3. Hostname: 157.245.206.124
4. Port: 3306
5. Username: admin_aqil
6. Password: admin_aqil (click "Store in Vault")
7. Default Schema: admin_railway
8. Test Connection & Save

### 5. Using DBeaver (Universal Database Tool)
1. New Database Connection → MySQL
2. Server Host: 157.245.206.124
3. Port: 3306
4. Database: admin_railway
5. Username: admin_aqil
6. Password: admin_aqil
7. Test Connection & Finish

### 6. Using HeidiSQL (Windows)
1. Session manager → New
2. Network type: MySQL (TCP/IP)
3. Hostname: 157.245.206.124
4. User: admin_aqil
5. Password: admin_aqil
6. Port: 3306
7. Database: admin_railway

### 7. Using TablePlus
1. Create new connection → MySQL
2. Host: 157.245.206.124
3. Port: 3306
4. User: admin_aqil
5. Password: admin_aqil
6. Database: admin_railway
7. Name: WhatsApp Multi-Device
8. Test & Save

### 8. Using Sequel Pro (Mac)
1. New Connection
2. MySQL Host: 157.245.206.124
3. Username: admin_aqil
4. Password: admin_aqil
5. Database: admin_railway
6. Port: 3306
7. Connect

## Connection String Formats

### MySQL URI Format
```
mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway
```

### JDBC Format (for Java)
```
jdbc:mysql://157.245.206.124:3306/admin_railway?user=admin_aqil&password=admin_aqil
```

### Node.js (mysql2)
```javascript
const mysql = require('mysql2');
const connection = mysql.createConnection({
  host: '157.245.206.124',
  port: 3306,
  user: 'admin_aqil',
  password: 'admin_aqil',
  database: 'admin_railway'
});
```

### PHP PDO
```php
$dsn = 'mysql:host=157.245.206.124;port=3306;dbname=admin_railway';
$pdo = new PDO($dsn, 'admin_aqil', 'admin_aqil');
```

### Go (go-sql-driver/mysql) - Current Implementation
```go
// MySQL 5.7 optimized connection string
dsn := "admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci"
db, err := sql.Open("mysql", dsn)

// MySQL 5.7 specific connection pool settings
db.SetMaxOpenConns(25)     // Optimal for MySQL 5.7
db.SetMaxIdleConns(5)      // Conservative for 5.7
db.SetConnMaxLifetime(5 * time.Minute) // Prevents connection timeouts
```

## MySQL 5.7 JSON Operations

### Working with JSON Fields

MySQL 5.7 introduced native JSON support, which this system uses extensively:

```sql
-- Query JSON fields in flows
SELECT id, name, JSON_EXTRACT(nodes, '$[0].type') as first_node_type 
FROM chatbot_flows_nodepath;

-- Update JSON nodes array
UPDATE chatbot_flows_nodepath 
SET nodes = JSON_ARRAY(
  JSON_OBJECT('id', 'start', 'type', 'message', 'data', JSON_OBJECT('text', 'Welcome!'))
) 
WHERE id = 'flow_001';


-- Search within JSON arrays (MySQL 5.7.8+)
SELECT * FROM chatbot_flows_nodepath 
WHERE JSON_SEARCH(nodes, 'one', 'message', NULL, '$[*].type') IS NOT NULL;
```

### MySQL 5.7 JSON Functions Used:
- `JSON_EXTRACT()` - Extract values from JSON
- `JSON_OBJECT()` - Create JSON objects
- `JSON_ARRAY()` - Create JSON arrays
- `JSON_SEARCH()` - Search within JSON documents
- `JSON_SET()` - Update JSON values
- `JSON_INSERT()` - Insert into JSON documents

## Database Operations Examples

### Direct SQL Operations

#### SELECT Operations
```sql
-- Get all chatbot flows
SELECT * FROM chatbot_flows_nodepath;


-- Get recent WhatsApp messages
SELECT * FROM whatsapp_messages ORDER BY created_at DESC LIMIT 10;

-- Get campaign analytics
SELECT * FROM message_analytics WHERE created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY);
```

#### INSERT Operations
```sql
-- Insert new chatbot flow
INSERT INTO chatbot_flows_nodepath (id, name, description, nodes, edges) 
VALUES ('flow_001', 'Welcome Flow', 'Initial welcome sequence', '[]', '[]');

-- Insert new lead
INSERT INTO chatbot_leads (phone_number, name, source, created_at) 
VALUES ('+1234567890', 'John Doe', 'website', NOW());
```

#### UPDATE Operations
```sql
-- Update flow configuration
UPDATE chatbot_flows_nodepath 
SET nodes = '[{"id":"start","type":"message"}]', updated_at = NOW() 
WHERE id = 'flow_001';

```

#### DELETE Operations
```sql

-- Delete test flows
DELETE FROM chatbot_flows_nodepath WHERE name LIKE 'Test%';
```

## System Configuration

### Environment Variables (.env)
```env
# Database Configuration
DATABASE_URL=admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local

# Legacy Vite Database Configuration (keep for compatibility)
VITE_DB_HOST=157.245.206.124
VITE_DB_NAME=admin_railway
VITE_DB_USER=admin_aqil
VITE_DB_PASSWORD=admin_aqil
VITE_DB_PORT=3306
```

### Go Configuration (internal/config/config.go)
```go
// Database configuration with fallback values
MySQLHost:     getEnv("MYSQL_HOST", "157.245.206.124"),
MySQLPort:     getEnvAsInt("MYSQL_PORT", 3306),
MySQLUser:     getEnv("MYSQL_USER", "admin_aqil"),
MySQLPassword: getEnv("MYSQL_PASSWORD", "admin_aqil"),
MySQLDatabase: getEnv("MYSQL_DATABASE", "admin_railway"),
```

## Testing Database Connection

Use the included test script:
```bash
go run test-db-connection.go
```

This will:
- Test basic connectivity
- Verify MySQL version
- List all available tables
- Confirm the connection is working properly

## Security Notes

⚠️ **Important Security Considerations:**

1. **Production Environment**: These credentials are for the production database
2. **Access Control**: Ensure proper network security and access controls
3. **Backup Strategy**: Implement regular database backups
4. **Connection Pooling**: The system uses connection pooling (max 25 connections)
5. **SSL/TLS**: Consider enabling SSL for production connections

## Troubleshooting

### Common Issues

1. **Connection Timeout**
   - Check network connectivity
   - Verify firewall settings
   - Ensure MySQL server is running

2. **Authentication Failed**
   - Verify username and password
   - Check user permissions in MySQL

3. **Database Not Found**
   - Confirm database name is correct
   - Verify user has access to the database

### Connection Pool Settings (MySQL 5.7 Optimized)

```go
// Current configuration in database.go - optimized for MySQL 5.7
db.SetMaxOpenConns(25)     // Maximum open connections (safe for 5.7)
db.SetMaxIdleConns(5)      // Maximum idle connections (conservative)
db.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime (prevents 5.7 timeouts)
```

**MySQL 5.7 Connection Pool Notes:**
- MySQL 5.7 has a default `max_connections` of 151
- Our pool of 25 connections is well within safe limits
- Connection lifetime prevents MySQL 5.7's `wait_timeout` issues
- Idle connection limit conserves MySQL 5.7 resources

## Database Schema

The system automatically creates and maintains the following core tables:

- **chatbot_flows_nodepath**: Flow configurations with JSON nodes and edges


Additional tables are managed by the WhatsApp integration and other system components.

---

## ⚠️ Important MySQL Version Notice

**Database Version**: MySQL 5.7.44  
**Compatibility**: This system is specifically designed for MySQL 5.7.x

### MySQL 5.7 Specific Considerations:

1. **JSON Support**: MySQL 5.7 introduced native JSON data type support
   - Used extensively in `chatbot_flows_nodepath.nodes` and `chatbot_flows_nodepath.edges`


2. **Character Set**: UTF8MB4 with unicode collation
   - All tables use `utf8mb4_unicode_ci` collation
   - Supports full Unicode including emojis (important for WhatsApp)

3. **SQL Mode Compatibility**: 
   - Ensure `sql_mode` is compatible with application queries
   - Default MySQL 5.7 modes should work fine

4. **Performance Optimizations**:
   - Connection pooling configured for MySQL 5.7
   - Indexes optimized for 5.7 query planner

5. **Migration Compatibility**:
   - All migrations tested on MySQL 5.7.44
   - JSON functions used are 5.7 compatible

---

**Last Updated**: January 2025  
**Database Version**: MySQL 5.7.44  
**Connection Status**: ✅ Verified Working