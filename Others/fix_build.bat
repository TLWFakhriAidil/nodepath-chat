@echo off
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix build error: Update GetDevice to use models.UserDevice"
git push origin main
echo Build fix pushed successfully!
pause