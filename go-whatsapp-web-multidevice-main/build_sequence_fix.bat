@echo off
echo === Sequence Fix Build Script ===
echo.
echo Building with sequence trigger fix...
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
go build -o whatsapp_sequence_fixed.exe ./src
if %errorlevel% neq 0 (
    echo Build failed! Please check for errors.
    pause
    exit /b 1
)
echo Build successful!
echo.
echo The fixed executable is: whatsapp_sequence_fixed.exe
echo Run it with: whatsapp_sequence_fixed.exe rest
pause
