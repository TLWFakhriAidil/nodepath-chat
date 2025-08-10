@echo off
echo Cleaning up project and fixing build errors
echo ==========================================

echo.
echo Removing problematic files...
del check_campaign_issue.go 2>nul
del device_manager_enhancement.go 2>nul

echo.
echo Installing missing dependency...
go get github.com/lib/pq

echo.
echo Building without CGO...
set CGO_ENABLED=0
go build -o whatsapp.exe

echo.
echo Build status:
if exist whatsapp.exe (
    echo SUCCESS! whatsapp.exe built successfully.
    echo.
    echo File size:
    dir whatsapp.exe | findstr whatsapp.exe
) else (
    echo FAILED! Build did not complete successfully.
)

echo.
echo Ready to push to GitHub:
echo   git push origin main
echo.
pause
