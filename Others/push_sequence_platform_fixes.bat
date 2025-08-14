@echo off
echo ========================================
echo Pushing Sequence Fix and Platform Logging
echo ========================================

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "Fix sequence activation logic & enhance platform API logging

FIXES:
1. Sequence Activation Fix:
   - Changed from MIN(current_step) to MIN(next_trigger_time)
   - Now activates steps based on scheduled time, not step number
   - Supports flexible scheduling (e.g., Step 3 before Step 2)
   - Respects trigger_delay_hours properly

2. Platform API Logging Enhancement:
   - Added comprehensive logging for Wablas and Whacenter
   - Includes request/response details with timing
   - Shows success/failure with visual indicators
   - Helps debug authentication and API errors
   - Performance monitoring with response times

Files Changed:
- src/usecase/sequence_trigger_processor.go
- src/pkg/external/platform_sender.go
- src/infrastructure/broadcast/whatsapp_message_sender.go

Documentation:
- SEQUENCE_ACTIVATION_FIX.md
- PLATFORM_LOGGING_ENHANCEMENT.md"

REM Push to main branch
echo.
echo Pushing to GitHub main branch...
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause
