@echo off
echo Applying sequence modal date filter fix...

REM Backup the original file
copy "src\ui\rest\app.go" "src\ui\rest\app.go.backup_%date:~10,4%%date:~4,2%%date:~7,2%_%time:~0,2%%time:~3,2%.bak"

echo.
echo ===================================
echo MANUAL FIX REQUIRED for app.go
echo ===================================
echo.
echo In src/ui/rest/app.go, find the GetSequenceStepLeads function (around line 4906)
echo.
echo 1. After line 4911 (status := c.Query("status", "all")), add:
echo.
echo    // Get date filters from query params
echo    startDate := c.Query("start_date")
echo    endDate := c.Query("end_date")
echo.
echo 2. After line 4947 (args := []interface{}{sequenceId, deviceId, stepId, session.UserID}), add:
echo.
echo    log.Printf("GetSequenceStepLeads - Sequence: %%s, Device: %%s, Step: %%s, Status: %%s, DateRange: %%s to %%s",
echo        sequenceId, deviceId, stepId, status, startDate, endDate)
echo.
echo 3. After the status filter conditions (around line 4957), before the ORDER BY clause, add:
echo.
echo    // Add date filter if provided
echo    if startDate != "" ^&^& endDate != "" {
echo        query += ` AND DATE(bm.sent_at) BETWEEN ? AND ?`
echo        args = append(args, startDate, endDate)
echo    } else if startDate != "" {
echo        query += ` AND DATE(bm.sent_at) ^>= ?`
echo        args = append(args, startDate)
echo    } else if endDate != "" {
echo        query += ` AND DATE(bm.sent_at) ^<= ?`
echo        args = append(args, endDate)
echo    }
echo.
echo ===================================
echo.

REM Backup dashboard.html
copy "src\views\dashboard.html" "src\views\dashboard.html.backup_%date:~10,4%%date:~4,2%%date:~7,2%_%time:~0,2%%time:~3,2%.bak"

echo ===================================
echo MANUAL FIX REQUIRED for dashboard.html
echo ===================================
echo.
echo In src/views/dashboard.html, find the showSequenceStepLeadDetails function (around line 7358)
echo.
echo After the line that builds the URL:
echo    let url = `/api/sequences/${currentSequenceForReport.id}/device/${deviceId}/step/${stepId}/leads?status=${status}`;
echo.
echo Add these lines:
echo.
echo    // Get the current date filters from the sequence summary
echo    const startDate = document.getElementById('sequenceStartDate').value;
echo    const endDate = document.getElementById('sequenceEndDate').value;
echo    
echo    if (startDate) {
echo        url += `^&start_date=${startDate}`;
echo    }
echo    if (endDate) {
echo        url += `^&end_date=${endDate}`;
echo    }
echo.
echo    console.log('Fetching sequence step leads with URL:', url);
echo.
echo ===================================
echo.
echo Please apply these changes manually, then:
echo 1. Build the application
echo 2. Test the fix
echo 3. Commit and push to GitHub
echo.
pause
