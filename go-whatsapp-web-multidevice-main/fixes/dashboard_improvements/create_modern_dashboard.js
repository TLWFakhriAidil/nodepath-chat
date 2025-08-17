const fs = require('fs');
const path = require('path');

console.log('Creating modern dashboard design...');

const dashboardPath = path.join(__dirname, '../../src/views/dashboard.html');
let dashboardContent = fs.readFileSync(dashboardPath, 'utf8');

// Add modern dashboard styles
const modernStyles = `
        /* Modern Dashboard Styles */
        .dashboard-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 2rem 0;
            margin: -1rem -1rem 2rem -1rem;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        
        .stat-card {
            background: white;
            border-radius: 16px;
            padding: 1.5rem;
            box-shadow: 0 2px 8px rgba(0,0,0,0.08);
            transition: all 0.3s;
            border: 1px solid #f0f0f0;
            height: 100%;
        }
        
        .stat-card:hover {
            transform: translateY(-4px);
            box-shadow: 0 8px 24px rgba(0,0,0,0.12);
        }
        
        .stat-number {
            font-size: 2.5rem;
            font-weight: 700;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin: 0.5rem 0;
        }
        
        .stat-change {
            font-size: 0.875rem;
            display: flex;
            align-items: center;
            gap: 4px;
        }
        
        .stat-change.positive {
            color: #10b981;
        }
        
        .stat-change.negative {
            color: #ef4444;
        }
        
        .activity-feed {
            max-height: 400px;
            overflow-y: auto;
            padding-right: 0.5rem;
        }
        
        .activity-feed::-webkit-scrollbar {
            width: 6px;
        }
        
        .activity-feed::-webkit-scrollbar-track {
            background: #f1f1f1;
            border-radius: 3px;
        }
        
        .activity-feed::-webkit-scrollbar-thumb {
            background: #888;
            border-radius: 3px;
        }
        
        .activity-item {
            padding: 1rem;
            border-left: 3px solid #667eea;
            margin-bottom: 1rem;
            background: #f8f9fa;
            border-radius: 0 8px 8px 0;
            transition: all 0.2s;
        }
        
        .activity-item:hover {
            background: #e9ecef;
            transform: translateX(4px);
        }
        
        .activity-time {
            font-size: 0.75rem;
            color: #6c757d;
        }
        
        .chart-container {
            background: white;
            border-radius: 16px;
            padding: 1.5rem;
            box-shadow: 0 2px 8px rgba(0,0,0,0.08);
            height: 400px;
        }
        
        .quick-actions {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
            margin-bottom: 2rem;
        }
        
        .quick-action-btn {
            background: white;
            border: 2px solid #e9ecef;
            border-radius: 12px;
            padding: 1rem;
            text-align: center;
            transition: all 0.3s;
            cursor: pointer;
            text-decoration: none;
            color: inherit;
        }
        
        .quick-action-btn:hover {
            border-color: #667eea;
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.2);
        }
        
        .quick-action-btn i {
            font-size: 2rem;
            color: #667eea;
            margin-bottom: 0.5rem;
            display: block;
        }`;

// Add styles to the head section
if (!dashboardContent.includes('Modern Dashboard Styles')) {
    dashboardContent = dashboardContent.replace('</style>', modernStyles + '\n    </style>');
}

// Create the new dashboard header HTML
const dashboardHeaderHTML = `
                <!-- Modern Dashboard Header -->
                <div class="dashboard-header">
                    <div class="container">
                        <h1 class="mb-0">Welcome back! 👋</h1>
                        <p class="mb-0 opacity-75">Here's what's happening with your WhatsApp campaigns today</p>
                    </div>
                </div>`;

// Add the header after the time range section
dashboardContent = dashboardContent.replace(
    '<div class="tab-pane fade show active" id="dashboard" role="tabpanel">',
    '<div class="tab-pane fade show active" id="dashboard" role="tabpanel">' + dashboardHeaderHTML
);

fs.writeFileSync(dashboardPath, dashboardContent, 'utf8');
console.log('Modern dashboard design applied successfully!');
