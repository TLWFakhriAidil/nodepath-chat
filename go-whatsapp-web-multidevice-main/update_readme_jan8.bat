@echo off
echo Updating README with latest fixes...

REM Commit and push
git add README.md
git commit -m "Update README with January 8, 2025 schema and query fixes

- Document successful sequence trigger processing
- Add details about schema mismatch resolution
- Update status to show system is working properly
- Include query optimization changes"

git push origin main

echo README updated and pushed successfully!
pause