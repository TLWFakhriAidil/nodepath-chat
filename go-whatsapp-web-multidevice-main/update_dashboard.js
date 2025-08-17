const fs = require('fs');
const path = require('path');

// Read the dashboard.html file
const filePath = path.join(__dirname, 'src', 'views', 'dashboard.html');
let content = fs.readFileSync(filePath, 'utf8');

// 1. Remove Reset WhatsApp Session menu item
const resetMenuPattern = /(\s*<li><hr class="dropdown-divider"><\/li>\s*<li><a class="dropdown-item text-warning"[^>]*onclick="resetDevice[^>]*>[^<]*<i class="bi bi-arrow-counterclockwise[^>]*><\/i>Reset WhatsApp Session\s*<\/a><\/li>)/g;
content = content.replace(resetMenuPattern, '');

// 2. Replace the logout function
const oldLogoutStart = '        // Logout Device\n        function logoutDevice(deviceId) {';
const oldLogoutEnd = '        }';

// Find the start and end of the function
const startIndex = content.indexOf(oldLogoutStart);
if (startIndex !== -1) {
    // Find the closing brace for this function
    let braceCount = 0;
    let endIndex = startIndex + oldLogoutStart.length;
    let inFunction = false;
    
    for (let i = endIndex; i < content.length; i++) {
        if (content[i] === '{') {
            braceCount++;
            inFunction = true;
        } else if (content[i] === '}') {
            braceCount--;
            if (inFunction && braceCount === 0) {
                endIndex = i + 1;
                break;
            }
        }
    }
    
    // Read the new function
    const newFunction = fs.readFileSync('new_logout_function.js', 'utf8');
    
    // Replace the old function with the new one
    content = content.substring(0, startIndex) + newFunction + content.substring(endIndex);
}

// 3. Also remove the resetDevice function since it's no longer needed
const resetFunctionPattern = /\s*\/\/ Reset Device WhatsApp Session\s*function resetDevice\(deviceId\) \{[\s\S]*?\n\s*\}\s*\}\s*\)\s*;?\s*\}/g;
content = content.replace(resetFunctionPattern, '');

// Write the updated content back
fs.writeFileSync(filePath, content, 'utf8');

console.log('Dashboard.html has been updated successfully!');
console.log('- Removed Reset WhatsApp Session menu item');
console.log('- Updated logout function to include session reset');
console.log('- Removed unused resetDevice function');
