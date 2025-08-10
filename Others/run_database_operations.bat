@echo off
echo Installing required Python packages...
pip install psycopg2-binary pymysql
echo.
echo Running database operations...
python database_operations.py
pause
