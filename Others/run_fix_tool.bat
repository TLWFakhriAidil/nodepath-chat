@echo off
echo ====================================
echo   SEQUENCE STEPS FIX TOOL
echo ====================================

echo.
echo This tool will fix the "Invalid Date" issue in your sequence steps
echo.

if not defined DATABASE_URL (
    echo ERROR: DATABASE_URL environment variable not set
    echo Please set your DATABASE_URL and run this script again
    echo.
    echo Example:
    echo set DATABASE_URL=postgresql://user:pass@host:port/database
    echo.
    pause
    exit /b 1
)

echo Database URL found: %DATABASE_URL%
echo.
echo Building fix tool...
go build -o sequence_fix_tool.exe sequence_fix_tool.go

if not exist sequence_fix_tool.exe (
    echo ERROR: Failed to build fix tool
    echo Make sure Go is installed and try again
    pause
    exit /b 1
)

echo ✓ Fix tool built successfully
echo.
echo Running sequence fix...
sequence_fix_tool.exe

echo.
echo Cleaning up...
del sequence_fix_tool.exe 2>nul

echo.
echo ====================================
echo  NEXT STEPS:
echo ====================================
echo 1. Restart your Go application
echo 2. Refresh your browser page  
echo 3. Check if sequence steps now appear
echo.
pause
