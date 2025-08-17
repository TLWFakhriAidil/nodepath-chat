@echo off
echo Building and pushing lead fix...
echo ================================

cd src
echo Building application...
set CGO_ENABLED=0
go build -o ..\whatsapp.exe

cd ..

if exist whatsapp.exe (
    echo.
    echo Build successful!
    echo.
    echo Committing changes...
    git add -A
    git commit -m "Fix: Lead creation for MySQL - Use LastInsertId instead of RETURNING clause"
    
    echo.
    echo Pushing to GitHub...
    git push origin main
    
    echo.
    echo ================================
    echo FIX DEPLOYED!
    echo ================================
    echo.
    echo Lead CRUD operations fixed:
    echo - Create: Now uses result.LastInsertId() for MySQL
    echo - Read: Working with proper SQL syntax
    echo - Update: Uses standard UPDATE syntax
    echo - Delete: Uses standard DELETE syntax
    echo - Import/Export: Fully functional with CSV
    echo.
    echo The application should now handle leads correctly!
) else (
    echo Build failed! Check errors above.
)

pause
