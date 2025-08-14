const fs = require('fs');
const path = require('path');

console.log('Adding navigation to all pages...');

// Navigation HTML component
const navigationHTML = `
    <!-- Navigation Bar -->
    <div class="navigation-bar bg-light py-2 px-3 border-bottom">
        <div class="container d-flex justify-content-between align-items-center">
            <div>
                <button class="btn btn-sm btn-outline-secondary" onclick="history.back()">
                    <i class="bi bi-arrow-left"></i> Back
                </button>
                <button class="btn btn-sm btn-outline-primary ms-2" onclick="window.location.href='/dashboard'">
                    <i class="bi bi-house"></i> Home
                </button>
            </div>
            <div class="breadcrumb mb-0">
                <span class="text-muted">You are here: </span>
                <span id="currentPage">Dashboard</span>
            </div>
        </div>
    </div>`;

// Navigation script
const navigationScript = `
    <script>
        // Set current page in breadcrumb
        document.addEventListener('DOMContentLoaded', function() {
            const pageName = document.title.split(' - ')[0] || 'Dashboard';
            const currentPageElement = document.getElementById('currentPage');
            if (currentPageElement) {
                currentPageElement.textContent = pageName;
            }
        });
    </script>`;

// List of HTML files to update
const htmlFiles = [
    'dashboard.html',
    'sequences.html',
    'sequence_detail.html',
    'device_actions.html',
    'device_leads.html',
    'whatsapp_web.html'
];

htmlFiles.forEach(file => {
    const filePath = path.join(__dirname, '../../src/views/', file);
    
    if (fs.existsSync(filePath)) {
        let content = fs.readFileSync(filePath, 'utf8');
        
        // Add navigation after navbar if not already present
        if (!content.includes('navigation-bar')) {
            // Find the end of navbar
            const navbarEndRegex = /<\/nav>\s*(?=<div|<main|<!--)/;
            
            if (navbarEndRegex.test(content)) {
                content = content.replace(navbarEndRegex, '</nav>' + navigationHTML);
            } else {
                // If no navbar, add after body tag
                content = content.replace(/<body[^>]*>/, '$&' + navigationHTML);
            }
            
            // Add navigation script before closing body tag
            if (!content.includes('currentPage')) {
                content = content.replace('</body>', navigationScript + '\n</body>');
            }
            
            fs.writeFileSync(filePath, content, 'utf8');
            console.log(`✓ Navigation added to ${file}`);
        } else {
            console.log(`✓ Navigation already exists in ${file}`);
        }
    } else {
        console.log(`✗ File not found: ${file}`);
    }
});

console.log('Navigation added to all pages!');
