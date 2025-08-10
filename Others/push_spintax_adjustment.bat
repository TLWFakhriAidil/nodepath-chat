@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Reduce homoglyph replacement from 10% to 5%" -m "- Changed applyHomoglyphs percentage from 0.10 to 0.05" -m "- This reduces character replacements with look-alikes from 10% to 5%" -m "- Makes messages more readable while still providing anti-spam variation"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
