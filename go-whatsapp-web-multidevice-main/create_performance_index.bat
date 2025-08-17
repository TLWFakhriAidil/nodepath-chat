@echo off
echo Creating performance indexes for broadcast_messages...
echo.

python -c "import pymysql; conn = pymysql.connect(host='159.89.198.71', port=3306, user='admin_aqil', password='admin_aqil', database='admin_railway'); cursor = conn.cursor(); cursor.execute('CREATE INDEX IF NOT EXISTS idx_broadcast_optimize ON broadcast_messages(status, device_id, scheduled_at)'); conn.commit(); print('Index created successfully!'); cursor.close(); conn.close()"

echo.
echo Done!
pause
