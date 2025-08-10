@echo off
echo Fixing sequence creation issues...

REM Compile the project
echo Building application...
cd src
go build -o whatsapp.exe
if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)

REM Copy to root
copy whatsapp.exe ..\ /Y

echo Build complete!
echo Please restart the application to apply changes.
pause
