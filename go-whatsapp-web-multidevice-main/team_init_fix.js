        // Initialize on page load
        document.addEventListener('DOMContentLoaded', function() {
            // Simple initialization - middleware handles auth
            console.log('Team Dashboard initialized');
            
            // Initialize components
            initializeDashboard();
            setDefaultDates();
            updateCurrentTime();
            setInterval(updateCurrentTime, 60000);
            
            // Load team member info
            fetch('/api/team-member/info', {
                credentials: 'include'
            })
            .then(response => response.json())
            .then(data => {
                if (data.success && data.member) {
                    document.getElementById('teamMemberName').textContent = data.member.username;
                }
            })
            .catch(error => console.error('Error loading team info:', error));
            
            // Setup tab listeners
            setupTabListeners();
            
            // Load initial dashboard data
            setTimeout(() => {
                loadDashboardData();
            }, 100);
        });
        
        // Set default date range
        function setDefaultDates() {
            const endDate = new Date();
            const startDate = new Date();
            startDate.setDate(startDate.getDate() - 7);
            
            document.getElementById('startDate').value = startDate.toISOString().split('T')[0];
            document.getElementById('endDate').value = endDate.toISOString().split('T')[0];
        }
