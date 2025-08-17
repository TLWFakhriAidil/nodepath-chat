@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo ========================================
echo Finding the first commit date...
echo ========================================

REM Get the first commit date
for /f "tokens=*" %%i in ('git log --reverse --pretty^=format:"%%ad" --date^=short -1') do set FIRST_DATE=%%i
echo First commit date: %FIRST_DATE%

echo.
echo ========================================
echo All commits from the first week:
echo ========================================

REM Show commits from the first 7 days
git log --reverse --pretty=format:"%%h - %%ad - %%an : %%s" --date=short --since="%FIRST_DATE%" --until="%FIRST_DATE% + 7 days"

echo.
echo.
echo ========================================
echo Alternative: First 50 commits (chronological order):
echo ========================================
git log --reverse --pretty=format:"%%h - %%ad - %%an : %%s" --date=short -50

echo.
pause
