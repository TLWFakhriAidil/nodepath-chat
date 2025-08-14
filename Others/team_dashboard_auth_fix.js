// Global flag to prevent multiple redirects
let isRedirecting = false;

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    // Clear any stale redirect flags when successfully on dashboard
    if (window.location.pathname === '/team-dashboard') {
        sessionStorage.removeItem('redirecting');
    }
    
    // Load team member info first
    loadTeamMemberInfo();
});

// Load team member info
async function loadTeamMemberInfo() {
    // Skip if already redirecting
    if (isRedirecting || sessionStorage.getItem('redirecting')) {
        return;
    }
    
    try {
        const response = await fetch('/api/team-member/info', {
            credentials: 'include'
        });
        
        if (!response.ok) {
            if (response.status === 401 && !isRedirecting) {
                isRedirecting = true;
                sessionStorage.setItem('redirecting', 'true');
                console.error('Authentication error - redirecting to login');
                window.location.href = '/team-login';
            }
            return;
        }
        
        const data = await response.json();
        if (data.member) {
            document.getElementById('teamMemberName').textContent = data.member.username;
            
            // Only after successful auth, initialize everything else
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

// Wrap all other API calls to check auth state first
async function makeAuthenticatedRequest(url, options = {}) {
    // Don't make requests if redirecting
    if (isRedirecting || sessionStorage.getItem('redirecting')) {
        return null;
    }
    
    const response = await fetch(url, {
        ...options,
        credentials: 'include'
    });
    
    if (!response.ok) {
        if (response.status === 401 && !isRedirecting) {
            // Don't redirect from here, let loadTeamMemberInfo handle it
            console.error('Auth error in request to:', url);
            return null;
        }
        throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    return response;
}
