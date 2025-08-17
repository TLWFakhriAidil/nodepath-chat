import os
import re
import shutil
from datetime import datetime

def backup_file(filepath):
    """Create a backup of the file"""
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    backup_path = f"{filepath}.backup_{timestamp}"
    shutil.copy2(filepath, backup_path)
    print(f"Backed up: {filepath} -> {backup_path}")
    return backup_path

def fix_go_file():
    """Fix the GetSequenceStepLeads function in app.go"""
    filepath = "src/ui/rest/app.go"
    
    if not os.path.exists(filepath):
        print(f"Error: {filepath} not found!")
        return False
    
    # Backup the file
    backup_file(filepath)
    
    # Read the file
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Find the GetSequenceStepLeads function
    pattern = r'(func \(handler \*App\) GetSequenceStepLeads\(c \*fiber\.Ctx\) error \{[^}]*?status := c\.Query\("status", "all"\))'
    match = re.search(pattern, content, re.DOTALL)
    
    if match:
        # Add date filter query params
        replacement = match.group(1) + '\n\n\t// Get date filters from query params\n\tstartDate := c.Query("start_date")\n\tendDate := c.Query("end_date")'
        content = content.replace(match.group(1), replacement)
        
        # Add logging after args initialization
        pattern2 = r'(args := \[\]interface\{\}\{sequenceId, deviceId, stepId, session\.UserID\})'
        match2 = re.search(pattern2, content)
        if match2:
            replacement2 = match2.group(1) + '\n\n\tlog.Printf("GetSequenceStepLeads - Sequence: %s, Device: %s, Step: %s, Status: %s, DateRange: %s to %s",\n\t\tsequenceId, deviceId, stepId, status, startDate, endDate)'
            content = content.replace(match2.group(1), replacement2)
        
        # Add date filter conditions before ORDER BY
        pattern3 = r'(\}\s*\}\s*\n\s*query \+= ` ORDER BY bm\.sent_at DESC`)'
        match3 = re.search(pattern3, content)
        if match3:
            date_filter = '''
\t// Add date filter if provided
\tif startDate != "" && endDate != "" {
\t\tquery += ` AND DATE(bm.sent_at) BETWEEN ? AND ?`
\t\targs = append(args, startDate, endDate)
\t} else if startDate != "" {
\t\tquery += ` AND DATE(bm.sent_at) >= ?`
\t\targs = append(args, startDate)
\t} else if endDate != "" {
\t\tquery += ` AND DATE(bm.sent_at) <= ?`
\t\targs = append(args, endDate)
\t}
\t
\tquery += ` ORDER BY bm.sent_at DESC`'''
            content = content.replace(match3.group(0), date_filter)
        
        # Write the updated content
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        
        print(f"✅ Successfully updated {filepath}")
        return True
    else:
        print(f"❌ Could not find GetSequenceStepLeads function in {filepath}")
        return False

def fix_html_file():
    """Fix the showSequenceStepLeadDetails function in dashboard.html"""
    filepath = "src/views/dashboard.html"
    
    if not os.path.exists(filepath):
        print(f"Error: {filepath} not found!")
        return False
    
    # Backup the file
    backup_file(filepath)
    
    # Read the file
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Find the showSequenceStepLeadDetails function and the URL building line
    pattern = r'(let url = `/api/sequences/\$\{currentSequenceForReport\.id\}/device/\$\{deviceId\}/step/\$\{stepId\}/leads\?status=\$\{status\}`;)'
    match = re.search(pattern, content)
    
    if match:
        # Add date filter to URL
        date_filter_code = '''let url = `/api/sequences/${currentSequenceForReport.id}/device/${deviceId}/step/${stepId}/leads?status=${status}`;
    
    // Get the current date filters from the sequence summary
    const startDate = document.getElementById('sequenceStartDate').value;
    const endDate = document.getElementById('sequenceEndDate').value;
    
    if (startDate) {
        url += `&start_date=${startDate}`;
    }
    if (endDate) {
        url += `&end_date=${endDate}`;
    }

    console.log('Fetching sequence step leads with URL:', url);'''
        
        content = content.replace(match.group(0), date_filter_code)
        
        # Write the updated content
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        
        print(f"✅ Successfully updated {filepath}")
        return True
    else:
        print(f"❌ Could not find showSequenceStepLeadDetails URL building in {filepath}")
        return False

if __name__ == "__main__":
    print("Applying Sequence Modal Date Filter Fix...")
    print("=" * 50)
    
    # Fix Go file
    print("\n1. Fixing app.go...")
    go_success = fix_go_file()
    
    # Fix HTML file
    print("\n2. Fixing dashboard.html...")
    html_success = fix_html_file()
    
    print("\n" + "=" * 50)
    
    if go_success and html_success:
        print("\n✅ All fixes applied successfully!")
        print("\nNext steps:")
        print("1. Build the application: build_nocgo.bat")
        print("2. Test the fix")
        print("3. Commit and push to GitHub")
    else:
        print("\n❌ Some fixes failed. Please check the errors above.")
        print("You may need to apply the fixes manually.")
