@echo off
echo Fixing team dashboard tab display issue...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src

REM Build the application
echo Building application...
set CGO_ENABLED=0
go build -o whatsapp.exe

REM Check if build was successful
if exist whatsapp.exe (
    echo Build successful!
) else (
    echo Build failed!
    pause
    exit /b 1
)

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

REM Commit and push changes
echo Committing changes...
git add -A
git commit -m "Fix team dashboard tabs showing same content - fixed missing closing bracket"

echo Pushing to GitHub...
git push origin main

echo.
echo Fix complete!
pause
