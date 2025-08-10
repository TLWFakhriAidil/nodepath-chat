@echo off
echo Committing Sequence and Broadcast Manager Features...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Running go mod tidy...
cd src
go mod tidy
cd ..

echo.
echo Git status...
git status

echo.
echo Adding all changes...
git add .

echo.
echo Committing changes...
git commit -m "feat: Add message sequences and optimized broadcast manager

- Implement message sequences with niche-based auto-enrollment
- Add broadcast manager with device workers for 3,000+ devices
- Custom delay settings (min/max) per device
- Campaign triggers based on date and niche matching
- Worker pool system for simultaneous processing
- Queue-based message processing with retry logic
- Rate limiting per device with random delays
- Health monitoring and auto-restart for workers
- Sequence UI with drag-drop steps
- Support for text, image, video, document messages
- Individual progress tracking per contact
- Lead management with niche categorization
- Database migrations for all new tables
- REST API endpoints for sequence management
- Automatic triggers for campaigns and sequences
- Optimized for 200 users with 15 devices each"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ✅ Successfully pushed sequence and broadcast features to GitHub!
pause