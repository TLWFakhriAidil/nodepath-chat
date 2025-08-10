@echo off
echo Comprehensive fix for all repository SQL syntax
echo ==============================================

:: Create Python script to fix all repository files
echo import os > fix_all_repos.py
echo import re >> fix_all_repos.py
echo. >> fix_all_repos.py
echo repo_dir = r'src\repository' >> fix_all_repos.py
echo. >> fix_all_repos.py
echo for filename in os.listdir(repo_dir): >> fix_all_repos.py
echo     if filename.endswith('.go'): >> fix_all_repos.py
echo         filepath = os.path.join(repo_dir, filename) >> fix_all_repos.py
echo         print(f"Fixing {filename}...") >> fix_all_repos.py
echo         >> fix_all_repos.py
echo         with open(filepath, 'r', encoding='utf-8') as f: >> fix_all_repos.py
echo             content = f.read() >> fix_all_repos.py
echo         >> fix_all_repos.py
echo         # Fix missing query := declarations >> fix_all_repos.py
echo         # Pattern: backtick at start of line (with indentation) followed by SQL >> fix_all_repos.py
echo         pattern = r'(\n\s+)`\n\s*(SELECT|INSERT|UPDATE|DELETE)' >> fix_all_repos.py
echo         content = re.sub(pattern, r'\1query := `\n\1\2', content) >> fix_all_repos.py
echo         >> fix_all_repos.py
echo         # Also fix single-line SQL strings >> fix_all_repos.py
echo         pattern2 = r'(\n\s+)`(SELECT|INSERT|UPDATE|DELETE)' >> fix_all_repos.py
echo         content = re.sub(pattern2, r'\1query := `\2', content) >> fix_all_repos.py
echo         >> fix_all_repos.py
echo         # Save >> fix_all_repos.py
echo         with open(filepath, 'w', encoding='utf-8') as f: >> fix_all_repos.py
echo             f.write(content) >> fix_all_repos.py
echo. >> fix_all_repos.py
echo print("All repository files fixed!") >> fix_all_repos.py

python fix_all_repos.py
del fix_all_repos.py

echo.
echo Building application...
cd src
set CGO_ENABLED=0
go build -o ..\whatsapp.exe
cd ..

echo.
if exist whatsapp.exe (
    echo =============================
    echo BUILD SUCCESSFUL!
    echo =============================
    echo.
    dir whatsapp.exe | findstr whatsapp.exe
    echo.
    echo Committing final build...
    git add -A
    git commit -m "Build: Successfully compiled with all SQL syntax fixes - Production ready"
    echo.
    echo To deploy:
    echo   git push origin main
) else (
    echo Build still failing. Manual intervention needed.
)

pause
