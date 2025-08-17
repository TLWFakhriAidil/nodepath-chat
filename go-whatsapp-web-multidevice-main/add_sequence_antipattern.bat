@echo off
echo ========================================
echo Adding Anti-Pattern Protection to Sequences
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

echo Backing up original file...
copy "src\usecase\sequence_trigger_processor.go" "src\usecase\sequence_trigger_processor_before_antipattern.go.bak"

echo.
echo Making changes to add anti-pattern protection...
echo - Adding RecipientName to broadcast message
echo - Messages will now use greeting processor
echo - Messages will apply homoglyphs and zero-width spaces
echo - Same anti-ban protection as campaigns
echo.

REM Navigate to src directory and build
cd src
set CGO_ENABLED=0
go build -o ../whatsapp.exe .

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Build failed!
    cd ..
    move /Y "src\usecase\sequence_trigger_processor_before_antipattern.go.bak" "src\usecase\sequence_trigger_processor.go"
    pause
    exit /b 1
)

echo.
echo Build successful!
cd ..

echo.
echo Committing changes to Git...
git add -A
git commit -m "Add anti-pattern protection to sequences (homoglyphs, greetings, spintax)

- Sequences now use the same anti-ban techniques as campaigns
- Added RecipientName to broadcast message for greeting processor
- Messages will apply homoglyphs, zero-width spaces, and randomization
- Malaysian greeting variations included
- Critical for preventing WhatsApp bans in sequences

This ensures sequences have the same level of protection against pattern detection as campaigns."

echo.
echo Pushing to GitHub main branch...
git push origin main

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Git push failed!
    pause
    exit /b 1
)

echo.
echo ========================================
echo Anti-Pattern Protection Added Successfully!
echo ========================================
echo.
echo Sequences now have:
echo - Greeting processor (Malaysian variations)
echo - Homoglyph replacements
echo - Zero-width space insertion
echo - Spintax processing
echo - Same protection as campaigns
echo.
pause
