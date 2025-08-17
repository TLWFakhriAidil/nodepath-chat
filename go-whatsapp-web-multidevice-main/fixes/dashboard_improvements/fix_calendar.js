const fs = require('fs');
const path = require('path');

console.log('Fixing calendar with day labels and multiple campaigns...');

const dashboardPath = path.join(__dirname, '../../src/views/dashboard.html');
let dashboardContent = fs.readFileSync(dashboardPath, 'utf8');

// Update renderCalendar function
const updatedRenderCalendar = `function renderCalendar(date) {
    const year = date.getFullYear();
    const month = date.getMonth();
    const firstDay = new Date(year, month, 1).getDay();
    const daysInMonth = new Date(year, month + 1, 0).getDate();
    const monthNames = ['January', 'February', 'March', 'April', 'May', 'June', 
                       'July', 'August', 'September', 'October', 'November', 'December'];
    const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
    
    // Update month display
    document.getElementById('currentMonth').textContent = \`\${monthNames[month]} \${year}\`;
    
    // Clear calendar
    const calendar = document.getElementById('campaignCalendar');
    calendar.innerHTML = '';
    
    // Add day headers
    dayNames.forEach(day => {
        calendar.innerHTML += \`<div class="calendar-header">\${day}</div>\`;
    });
    
    // Add empty cells for days before month starts
    for (let i = 0; i < firstDay; i++) {
        calendar.innerHTML += '<div class="calendar-day other-month"></div>';
    }
    
    // Add days of month
    for (let day = 1; day <= daysInMonth; day++) {
        const dateStr = \`\${year}-\${String(month + 1).padStart(2, '0')}-\${String(day).padStart(2, '0')}\`;
        const campaignsForDay = campaigns.filter(c => c.campaign_date === dateStr);
        
        let dayHtml = \`
            <div class="calendar-day \${campaignsForDay.length > 0 ? 'has-campaign' : ''}" 
                 onclick="openCampaignModal('\${dateStr}')" 
                 data-date="\${dateStr}">
                <div class="calendar-date">\${day}</div>
        \`;
        
        if (campaignsForDay.length > 0) {
            dayHtml += '<div class="calendar-campaigns">';
            campaignsForDay.forEach((campaign, index) => {
                if (index < 3) { // Show max 3 campaigns
                    const time = campaign.scheduled_time ? campaign.scheduled_time.substring(0, 5) : 'All day';
                    dayHtml += \`
                        <div class="campaign-item" title="\${campaign.title}">
                            <span class="badge bg-primary" style="font-size: 10px;">\${time}</span>
                            <small style="display: block; font-size: 11px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">\${campaign.title}</small>
                        </div>
                    \`;
                }
            });
            if (campaignsForDay.length > 3) {
                dayHtml += \`<small class="text-muted" style="font-size: 10px;">+\${campaignsForDay.length - 3} more</small>\`;
            }
            dayHtml += '</div>';
        }
        
        dayHtml += '</div>';
        calendar.innerHTML += dayHtml;
    }
    
    // Fill remaining cells
    const totalCells = firstDay + daysInMonth;
    const remainingCells = 42 - totalCells; // 6 rows * 7 days
    for (let i = 0; i < remainingCells; i++) {
        calendar.innerHTML += '<div class="calendar-day other-month"></div>';
    }
}`;

// Replace the renderCalendar function
const calendarRegex = /function renderCalendar\(date\) {[\s\S]*?^}/m;
if (dashboardContent.match(calendarRegex)) {
    dashboardContent = dashboardContent.replace(calendarRegex, updatedRenderCalendar);
} else {
    // If function doesn't exist, add it
    dashboardContent = dashboardContent.replace(
        '// Campaign Functions',
        '// Campaign Functions\n\n' + updatedRenderCalendar + '\n'
    );
}

// Update calendar styles to support multiple campaigns
const calendarStyles = `
        .calendar-campaigns {
            margin-top: 4px;
        }
        
        .campaign-item {
            font-size: 11px;
            margin: 2px 0;
            padding: 2px 4px;
            background: rgba(255, 255, 255, 0.9);
            border-radius: 3px;
            display: flex;
            align-items: center;
            gap: 4px;
        }
        
        .campaign-item .badge {
            flex-shrink: 0;
            font-size: 9px !important;
            padding: 1px 4px;
        }
        
        .calendar-day.has-campaign {
            background: #e3f2fd;
            border: 1px solid var(--primary);
        }
        
        .calendar-header {
            background: var(--primary);
            color: white;
            padding: 10px;
            text-align: center;
            font-weight: 600;
            font-size: 14px;
        }`;

// Add styles if not present
if (!dashboardContent.includes('.calendar-campaigns')) {
    dashboardContent = dashboardContent.replace(
        '</style>',
        calendarStyles + '\n    </style>'
    );
}

fs.writeFileSync(dashboardPath, dashboardContent, 'utf8');
console.log('Calendar fixed with day labels and multiple campaign support!');
