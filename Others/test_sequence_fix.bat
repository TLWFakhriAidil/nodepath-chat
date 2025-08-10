@echo off
echo.
echo ========================================
echo 🔍 SEQUENCE STEPS FIX VERIFICATION
echo ========================================
echo.
echo Testing if the sequence steps fix worked...
echo.

REM Test the API endpoint
echo Testing API endpoint: http://localhost:3000/api/sequences
echo.

REM Use curl if available, otherwise use PowerShell
where curl >nul 2>&1
if %ERRORLEVEL% equ 0 (
    echo Using curl to test...
    curl -s http://localhost:3000/api/sequences > test_result.json
    
    if %ERRORLEVEL% equ 0 (
        echo ✅ API responded successfully!
        echo.
        echo Checking for step_count and steps data...
        
        REM Check if we have step counts > 0
        findstr /C:"step_count" test_result.json >nul
        if %ERRORLEVEL% equ 0 (
            echo ✅ Found step_count in response
            
            REM Show the step counts
            echo.
            echo Step counts found:
            findstr /C:"step_count" test_result.json
            
            REM Check if any step_count is > 0
            findstr /C:'"step_count": [1-9]' test_result.json >nul
            if %ERRORLEVEL% equ 0 (
                echo.
                echo 🎉 SUCCESS! Found sequences with steps (step_count > 0)
                echo.
                echo Checking for steps arrays...
                findstr /C:'"steps"' test_result.json >nul
                if %ERRORLEVEL% equ 0 (
                    echo ✅ Found steps arrays in response
                    echo.
                    echo 🎯 VERIFICATION RESULT: SEQUENCE STEPS FIX WORKED! 
                    echo.
                    echo Your sequences now have:
                    echo - step_count greater than 0
                    echo - Non-empty steps arrays
                    echo.
                    echo The issue has been resolved! 🎉
                ) else (
                    echo ⚠️  Steps arrays still missing from API response
                )
            ) else (
                echo ❌ All sequences still show step_count: 0
                echo.
                echo The fix may not have been applied correctly.
                echo Please try running the emergency fix again.
            )
        ) else (
            echo ⚠️  No step_count field found in response
        )
        
        echo.
        echo Full API response saved to: test_result.json
        echo.
        echo Would you like to see the full response? (Y/N):
        set /p show=""
        if /i "%show%"=="y" (
            echo.
            echo ========================================
            echo FULL API RESPONSE:
            echo ========================================
            type test_result.json
        )
        
    ) else (
        echo ❌ API request failed. Is the application running?
        echo Make sure the WhatsApp application is started and accessible at localhost:3000
    )
) else (
    echo curl not found, using PowerShell...
    powershell -Command "try { $response = Invoke-RestMethod -Uri 'http://localhost:3000/api/sequences' -Method Get; $response | ConvertTo-Json -Depth 10 | Out-File -FilePath 'test_result.json'; Write-Host '✅ API responded successfully!'; $stepCounts = ($response.results | Where-Object { $_.step_count -gt 0 }).Count; if ($stepCounts -gt 0) { Write-Host '🎉 SUCCESS! Found $stepCounts sequences with steps'; Write-Host 'The sequence steps fix worked!'; } else { Write-Host '❌ All sequences still show step_count: 0'; Write-Host 'The fix may not have been applied correctly.'; } } catch { Write-Host '❌ API request failed. Is the application running?'; }"
)

echo.
echo ========================================
echo Test completed!
echo ========================================
echo.

if exist test_result.json (
    echo Result saved to: test_result.json
    echo.
    echo Quick summary from the API response:
    echo.
    
    REM Show just the key fields
    findstr /C:"step_count" test_result.json 2>nul | head -5
    echo.
    
    echo For full details, check the test_result.json file.
)

echo.
pause
