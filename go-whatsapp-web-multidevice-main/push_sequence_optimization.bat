@echo off
echo Pushing Sequence Optimization Improvements to GitHub...

REM Add the improvements folder
git add improvements/

REM Add any modified files
git add src/usecase/sequence_trigger_processor.go

REM Commit with descriptive message
git commit -m "feat: Optimize sequence trigger processor for 3000 devices

- Create individual flow records per step with sequence_stepid tracking
- Remove retry logic - single attempt only for better performance
- Increase workers to 100 and batch size to 10,000
- Add smart device load balancing with scoring algorithm
- Implement random delays between min/max seconds
- Track device loads in device_load_balance table
- Respect sequence schedule times with 10-minute window
- Add comprehensive monitoring views
- Update processing_device_id and completed_at for each flow"

REM Push to main branch
git push origin main

echo.
echo Push complete! Check GitHub for the updates.
pause