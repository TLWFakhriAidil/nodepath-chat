@echo off
echo ========================================
echo DEPLOYING LEAD STATUS TARGETING
echo ========================================

echo.
echo Adding database migrations...
cd src
go run ../run_migration.go
cd ..

echo.
echo Committing lead status targeting feature...
git add -A
git commit -m "feat: Add lead status targeting for campaigns and sequences

- Added target_status field to campaigns (all, prospect, customer)
- Added target_status field to sequences (all, prospect, customer)
- Updated lead repository with GetLeadsByNicheAndStatus function
- Campaign/sequence now target leads by BOTH niche AND status
- Frontend updated with target status dropdown
- Supports comma-separated niches (EXSTART,ITADRESS)

Example use cases:
- Campaign for 'ITADRESS' niche targeting only 'prospect' status
- Sequence for 'EXSTART' niche targeting only 'customer' status
- Campaign for all niches targeting all statuses

Database changes:
- ALTER TABLE campaigns ADD COLUMN target_status VARCHAR(50) DEFAULT 'all'
- ALTER TABLE sequences ADD COLUMN target_status VARCHAR(50) DEFAULT 'all'"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================
echo DEPLOYMENT COMPLETE!
echo ========================================
echo.
echo New Features:
echo 1. Campaigns can target by niche AND status
echo 2. Sequences can target by niche AND status
echo 3. Lead with 'EXSTART,ITADRESS' will receive messages
echo    if campaign targets 'ITADRESS' niche
echo 4. Status options: all, prospect, customer
echo.
pause
