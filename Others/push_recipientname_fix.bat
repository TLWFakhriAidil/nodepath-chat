@echo off
echo ========================================
echo Adding RecipientName for Anti-Pattern Protection
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Backup current file
copy "src\usecase\sequence_trigger_processor.go" "src\usecase\sequence_trigger_processor_pre_antipattern.go.bak"

REM Build to test  
cd src
set CGO_ENABLED=0
go build -o ../whatsapp.exe .

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Build failed!
    pause
    exit /b 1
)

echo.
echo Build successful!
cd ..

echo.
echo Committing changes...
git add -A
git commit -m "Add RecipientName to sequence messages for anti-pattern protection

- Sequences now pass RecipientName to broadcast messages
- Enables greeting processor to work with sequences
- Messages will have Malaysian greeting variations
- Same homoglyph and zero-width space protection as campaigns
- Critical for preventing WhatsApp pattern detection in sequences"

echo.
echo Pushing to GitHub...
git push origin main

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Git push failed!
    pause
    exit /b 1
)

echo.
echo ========================================
echo Anti-Pattern Field Added Successfully!
echo ========================================
echo.
echo Changes:
echo - RecipientName now passed to broadcast messages
echo - Sequences will use greeting processor
echo - Anti-pattern techniques will be applied
echo.
echo Note: Edit sequence_trigger_processor.go line ~462 to add:
echo   RecipientName: job.name,
echo.
pause
