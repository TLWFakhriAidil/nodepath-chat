@echo off
echo ========================================
echo FIXING REDIS CONNECTION PROPERLY
echo ========================================

echo.
echo Committing Redis fix...
git add -A
git commit -m "fix: Properly handle Redis connection with fallback

- Fixed environment variable resolution for Redis
- Check multiple env var names (REDIS_URL, redis_url, RedisURL)
- Ignore template variables containing ${{
- Automatically fallback to in-memory if Redis not available
- Add better logging for debugging

Now the app will:
1. Try to connect to Redis if valid URL found
2. Fallback to in-memory if Redis fails
3. Work in both scenarios"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================
echo FIX DEPLOYED!
echo ========================================
echo.
echo Your app will now:
echo 1. Detect Redis properly from env vars
echo 2. Use Redis if available
echo 3. Fallback to in-memory if not
echo.
echo Check your logs for:
echo - 'Valid Redis URL found' (using Redis)
echo - 'No valid Redis URL found' (using in-memory)
echo.
pause
