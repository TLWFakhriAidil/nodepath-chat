@echo off
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix compilation errors and update README with optimization details"
git push origin main
echo.
echo Compilation fixes deployed!
echo.
echo Fixed:
echo - Removed unused variables (client, jid)
echo - Fixed FamilyName field (doesn't exist in ContactInfo)
echo - Removed unused context import
echo - Updated README with architecture details
pause
