@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo ========================================
echo First 20 commits in chronological order:
echo ========================================
git log --reverse --pretty=format:"%%h - %%ad (%%ar) - %%an : %%s" --date=iso-local -20

echo.
echo.
echo ========================================
echo Total number of commits in repository:
echo ========================================
git rev-list --count HEAD

echo.
echo ========================================
echo Date range of all commits:
echo ========================================
echo First commit:
git log --reverse --pretty=format:"%%ad - %%s" --date=iso-local -1
echo.
echo.
echo Latest commit:
git log --pretty=format:"%%ad - %%s" --date=iso-local -1

echo.
pause
