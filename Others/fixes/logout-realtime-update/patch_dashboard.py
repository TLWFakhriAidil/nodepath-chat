import os
import re

def patch_dashboard_html():
    """Patch dashboard.html to add DEVICE_LOGGED_OUT handler"""
    
    dashboard_path = "src/views/dashboard.html"
    
    # Read the file
    with open(dashboard_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Find the WebSocket switch statement
    # Look for the pattern where DEVICE_STATUS case is handled
    pattern = r"(case 'DEVICE_STATUS':.*?break;)"
    
    # The new case to add after DEVICE_STATUS
    new_case = """
                        case 'DEVICE_LOGGED_OUT':
                            // Update device status to offline when logged out
                            console.log('Device logged out:', data.result);
                            const loggedOutDeviceId = data.result?.deviceId;
                            if (loggedOutDeviceId) {
                                const device = devices.find(d => d.id === loggedOutDeviceId);
                                if (device) {
                                    device.status = 'offline';
                                    device.phone = '';
                                    device.jid = '';
                                    device.lastSeen = new Date().toISOString();
                                    renderDevices();
                                    
                                    // Show notification
                                    showAlert('warning', `Device ${device.name} has been logged out`);
                                }
                            }
                            break;"""
    
    # Check if DEVICE_LOGGED_OUT handler already exists
    if "case 'DEVICE_LOGGED_OUT':" in content:
        print("DEVICE_LOGGED_OUT handler already exists in dashboard.html")
        return False
    
    # Find and replace
    match = re.search(pattern, content, re.DOTALL)
    if match:
        # Insert the new case after DEVICE_STATUS case
        replacement = match.group(0) + new_case
        content = content.replace(match.group(0), replacement)
        
        # Backup original file
        backup_path = dashboard_path + ".backup"
        with open(backup_path, 'w', encoding='utf-8') as f:
            f.write(content)
        
        # Write updated content
        with open(dashboard_path, 'w', encoding='utf-8') as f:
            f.write(content)
        
        print(f"Successfully patched {dashboard_path}")
        print(f"Backup saved to {backup_path}")
        return True
    else:
        print("Could not find the WebSocket switch statement pattern")
        print("Please manually add the DEVICE_LOGGED_OUT handler")
        return False

if __name__ == "__main__":
    os.chdir(os.path.dirname(os.path.abspath(__file__)))
    os.chdir("../..")  # Go to project root
    patch_dashboard_html()
