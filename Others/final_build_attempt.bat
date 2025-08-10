@echo off
echo Final build attempt after fixing all syntax errors
echo =================================================

cd src
set CGO_ENABLED=0
echo Building application...
go build -o ..\whatsapp.exe

cd ..

echo.
if exist whatsapp.exe (
    echo =============================
    echo BUILD SUCCESSFUL!
    echo =============================
    echo.
    dir whatsapp.exe | findstr whatsapp.exe
    echo.
    echo Creating final commit...
    git add -A
    git commit -m "Fix: Resolved all Go syntax errors - Fixed reserved keywords in infrastructure/whatsapp"
    echo.
    echo =============================
    echo READY FOR DEPLOYMENT
    echo =============================
    echo.
    echo Run the following command to push to GitHub:
    echo   git push origin main
    echo.
    echo Then deploy to Railway or your preferred platform.
) else (
    echo Build failed. Checking remaining errors...
    cd src
    go build -o ..\whatsapp.exe 2^> ..\build_errors.txt
    cd ..
    echo.
    echo Error details saved to build_errors.txt
    type build_errors.txt
)

echo.
pause
