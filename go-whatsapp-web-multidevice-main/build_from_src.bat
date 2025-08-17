@echo off
echo Final build attempt with correct structure
echo =========================================

echo.
echo Building from the correct directory...
cd src
set CGO_ENABLED=0
go build -o ..\whatsapp.exe

cd ..

echo.
if exist whatsapp.exe (
    echo SUCCESS! Build completed.
    echo.
    dir whatsapp.exe | findstr whatsapp.exe
    echo.
    echo Ready to push to GitHub!
    echo Run: git push origin main
) else (
    echo Build failed. Checking for main.go...
    dir /s main.go
)

echo.
pause
