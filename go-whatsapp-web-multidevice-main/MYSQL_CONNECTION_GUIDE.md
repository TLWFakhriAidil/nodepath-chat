# MySQL Database Connection Guide

## Direct MySQL Connection Details

### Database Credentials
```
Host: 159.89.198.71
Port: 3306
Database: admin_railway
Username: admin_aqil
Password: admin_aqil
```

## Connection Methods

### 1. Using MySQL Command Line
```bash
mysql -h 159.89.198.71 -P 3306 -u admin_aqil -padmin_aqil admin_railway
```

### 2. Using phpMyAdmin
```
URL: http://159.89.198.71/phpmyadmin
Username: admin_aqil
Password: admin_aqil
```

### 3. Using Python (pymysql)
```python
import pymysql

connection = pymysql.connect(
    host='159.89.198.71',
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
3. Hostname: 159.89.198.71
4. Port: 3306
5. Username: admin_aqil
6. Password: admin_aqil (click "Store in Vault")
7. Default Schema: admin_railway
8. Test Connection & Save

### 5. Using DBeaver (Universal Database Tool)
1. New Database Connection → MySQL
2. Server Host: 159.89.198.71
3. Port: 3306
4. Database: admin_railway
5. Username: admin_aqil
6. Password: admin_aqil
7. Test Connection & Finish

### 6. Using HeidiSQL (Windows)
1. Session manager → New
2. Network type: MySQL (TCP/IP)
3. Hostname: 159.89.198.71
4. User: admin_aqil
5. Password: admin_aqil
6. Port: 3306
7. Database: admin_railway

### 7. Using TablePlus
1. Create new connection → MySQL
2. Host: 159.89.198.71
3. Port: 3306
4. User: admin_aqil
5. Password: admin_aqil
6. Database: admin_railway
7. Name: WhatsApp Multi-Device
8. Test & Save

### 8. Using Sequel Pro (Mac)
1. New Connection
2. MySQL Host: 159.89.198.71
3. Username: admin_aqil
4. Password: admin_aqil
5. Database: admin_railway
6. Port: 3306
7. Connect

### 9. Connection String Formats

**MySQL URI Format:**
```
mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway
```

**JDBC Format (for Java):**
```
jdbc:mysql://159.89.198.71:3306/admin_railway?user=admin_aqil&password=admin_aqil
```

**Node.js (mysql2):**
```javascript
const mysql = require('mysql2');
const connection = mysql.createConnection({
  host: '159.89.198.71',
  port: 3306,
  user: 'admin_aqil',
  password: 'admin_aqil',
  database: 'admin_railway'
});
```

**PHP PDO:**
```php
$dsn = 'mysql:host=159.89.198.71;port=3306;dbname=admin_railway';
$pdo = new PDO($dsn, 'admin_aqil', 'admin_aqil');
```

**Go (go-sql-driver/mysql):**
```go
dsn := "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway?parseTime=true"
db, err := sql.Open("mysql", dsn)
```

## Important Tables

### Main Tables:
- `broadcast_messages` - Message queue
- `campaigns` - Campaign definitions
- `sequences` - Sequence definitions
- `sequence_steps` - Sequence message templates
- `sequence_contacts` - Enrolled contacts
- `leads` - Contact database
- `user_devices` - WhatsApp devices
- `users` - System users

## Useful Queries

### Check pending messages:
```sql
SELECT COUNT(*) as total, status, DATE(scheduled_at) as date
FROM broadcast_messages 
WHERE DATE(created_at) >= CURDATE()
GROUP BY status, DATE(scheduled_at)
ORDER BY date, status;
```

### Check worker activity:
```sql
SELECT 
    processing_worker_id,
    COUNT(*) as messages,
    MIN(processing_started_at) as started,
    MAX(processing_started_at) as latest
FROM broadcast_messages 
WHERE processing_worker_id IS NOT NULL
AND DATE(created_at) = CURDATE()
GROUP BY processing_worker_id
ORDER BY started DESC
LIMIT 20;
```

### Check device status:
```sql
SELECT 
    id,
    device_name,
    status,
    platform,
    phone_number,
    last_seen
FROM user_devices
ORDER BY status DESC, last_seen DESC;
```

### Fix stuck messages:
```sql
UPDATE broadcast_messages 
SET scheduled_at = NOW()
WHERE status = 'pending'
AND scheduled_at < DATE_SUB(NOW(), INTERVAL 1 HOUR)
AND scheduled_at > DATE_SUB(NOW(), INTERVAL 48 HOUR)
LIMIT 500;
```

## Timezone Note
- Server timezone appears to be UTC
- Application uses +8 hours for Malaysia time
- All scheduled times are adjusted with DATE_ADD(NOW(), INTERVAL 8 HOUR)

## Backup Command
```bash
mysqldump -h 159.89.198.71 -P 3306 -u admin_aqil -padmin_aqil admin_railway > backup_$(date +%Y%m%d_%H%M%S).sql
```

---
Saved on: August 12, 2025
For: WhatsApp Multi-Device Broadcast System
