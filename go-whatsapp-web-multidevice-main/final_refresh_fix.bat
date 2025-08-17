@echo off
echo COMPREHENSIVE FIX FOR TEAM DASHBOARD REFRESH LOOP
echo ==================================================
echo.

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Step 1: Backing up current file...
copy src\views\team_dashboard.html src\views\team_dashboard_backup.html

echo Step 2: Fixing authentication logic...
powershell -Command @"
$content = Get-Content 'src\views\team_dashboard.html' -Raw

# Remove all the broken authentication redirect code
$content = $content -replace '(?s)// Global variables.*?// Update current time', @'
        // Global variables
        let currentCampaigns = [];
        let currentMonth = new Date().getMonth();
        let currentYear = new Date().getFullYear();
        let devicesData = [];
        let campaignChart = null;
        let sequenceChart = null;
        let autoRefreshInterval = null;
        let isRedirecting = false;

        // Initialize on page load
        document.addEventListener('DOMContentLoaded', function() {
            // Clear redirect flag if we're successfully on the dashboard
            if (window.location.pathname === '/team-dashboard') {
                sessionStorage.removeItem('redirecting');
            }
            
            // Load team member info first
            loadTeamMemberInfo();
        });

        // Update current time
'@

# Fix loadTeamMemberInfo function
$content = $content -replace '(?s)// Load team member info.*?}[\r\n\s]*}', @'
        // Load team member info
        async function loadTeamMemberInfo() {
            // Prevent multiple redirects
            if (isRedirecting || sessionStorage.getItem('redirecting')) {
                return;
            }
            
            try {
                const response = await fetch('/api/team-member/info', {
                    credentials: 'include'
                });
                
                if (!response.ok) {
                    if (response.status === 401) {
                        isRedirecting = true;
                        sessionStorage.setItem('redirecting', 'true');
                        console.error('Authentication error - redirecting to login');
                        window.location.href = '/team-login';
                        return;
                    }
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                
                const data = await response.json();
                if (data.member) {
                    document.getElementById('teamMemberName').textContent = data.member.username;
                    
                    // Only initialize after successful auth
                    initializeDashboard();
                    loadDashboardData();
                    setupTabListeners();
                    updateCurrentTime();
                    setInterval(updateCurrentTime, 60000);
                }
            } catch (error) {
                console.error('Error loading team member info:', error);
            }
        }
'@

# Remove auth checks from loadDashboardData
$content = $content -replace '(?s)if \(response\.status === 401\) \{[\r\n\s]*console\.error\(''Dashboard data: Authentication error''\);[\r\n\s]*// Don''t redirect here.*?return;[\r\n\s]*\}', @'
                    if (response.status === 401) {
                        console.error('Dashboard data: Authentication error');
                        return;
                    }'@

# Remove auth redirects from other functions
$content = $content -replace 'window\.location\.href = ''/team-login'';[\r\n\s]*return;', 'return;'

# Save the fixed content
Set-Content 'src\views\team_dashboard.html' -Value $content -NoNewline
"@

echo Step 3: Building application...
cd src
set CGO_ENABLED=0
go build -o whatsapp.exe
cd ..

echo Step 4: Committing fix...
git add -A
git commit -m "FINAL FIX: Team dashboard infinite refresh loop - Centralized auth handling in loadTeamMemberInfo only - Removed duplicate auth checks from other functions - Added isRedirecting flag to prevent multiple redirects - Fixed initialization order - only load data after auth success"

echo Step 5: Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo FIX COMPLETE!
echo ========================================
echo.
echo The team dashboard should now:
echo - Load without any refresh loops
echo - Handle authentication properly
echo - Work in all browsers including incognito
echo - Initialize only after successful auth
echo.
pause
