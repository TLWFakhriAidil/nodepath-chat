@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo === Pushing logging optimization to prevent crashes ===
echo.

git add src/infrastructure/broadcast/ultra_scale_redis_manager.go
git add src/usecase/campaign_trigger.go
git add src/usecase/optimized_campaign_trigger.go
git add src/usecase/sequence.go
git add src/repository/campaign_repository.go
git commit -m "perf: Remove excessive logging to prevent system overload

- Health check only logs when there are active workers
- Campaign triggers only log when campaigns are found
- Removed repetitive device status logging
- Timezone warning only shows once instead of every minute
- Removed verbose repository logging
- System now runs quietly unless there's actual work

This prevents log spam and reduces CPU/memory usage"

git push origin main

echo.
echo === Logging optimizations pushed! ===
echo.
echo The system will now run much quieter and more efficiently.
echo Only important events will be logged.
pause
