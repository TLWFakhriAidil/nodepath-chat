@echo off
echo Running sequence status column migration...
echo.

REM You need to set your database connection details here
set PGPASSWORD=your_password
set DB_HOST=localhost
set DB_PORT=5432
set DB_NAME=your_database
set DB_USER=your_username

REM Run the migration SQL
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -f database\migrations\add_sequence_status_column.sql

if %ERRORLEVEL% == 0 (
    echo.
    echo Migration completed successfully!
    echo Sequences table now has status column that supports active/inactive states.
) else (
    echo.
    echo Migration failed! Please check your database connection settings.
)

pause
