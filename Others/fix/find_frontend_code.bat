@echo off
echo Finding frontend JavaScript code...
echo.

echo === Searching for device report related code ===
findstr /s /i "device.report\|deviceReport\|showDevice\|totalLeads" *.go *.html *.js 2>nul

echo.
echo === Searching for modal code ===
findstr /s /i "modal\|showModal\|openModal" *.go *.html *.js 2>nul | findstr /i "device\|lead"

echo.
echo === Searching for click handlers ===
findstr /s /i "onclick\|addEventListener.*click" *.go *.html *.js 2>nul | findstr /i "lead\|total"

echo.
echo === Searching for API calls to device-report ===
findstr /s /i "device-report\|/api/campaigns/.*/device" *.go *.html *.js 2>nul

echo.
echo === Searching for embedded HTML/JS in Go files ===
findstr /s /i "<script>\|innerHTML\|<div.*modal" *.go 2>nul

pause