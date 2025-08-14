#!/usr/bin/env python3
"""
Fix campaign creation/editing issues in dashboard.html
This script fixes the "Cannot read properties of null (reading 'value')" error
"""

import re
import shutil
from datetime import datetime

def fix_dashboard_campaign_modal(file_path):
    """Fix the campaign modal issues in dashboard.html"""
    
    # Backup the original file
    backup_path = f"{file_path}.backup_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
    shutil.copy2(file_path, backup_path)
    print(f"Created backup: {backup_path}")
    
    # Read the file
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Fix 1: Update saveCampaign function to handle null campaignDate
    save_campaign_pattern = r'function saveCampaign\(\) \{[^}]+const campaignDate = document\.getElementById\(\'campaignDate\'\)\.value;'
    save_campaign_replacement = '''function saveCampaign() {
            const campaignId = document.getElementById('campaignId').value;
            let campaignDate = document.getElementById('campaignDate').value;
            
            // If campaignDate is null or empty, use today's date
            if (!campaignDate) {
                const today = new Date();
                campaignDate = today.toISOString().split('T')[0]; // Format: YYYY-MM-DD
                console.warn('Campaign date was empty, using today:', campaignDate);
                document.getElementById('campaignDate').value = campaignDate;
            }'''
    
    content = re.sub(save_campaign_pattern, save_campaign_replacement, content, flags=re.DOTALL)
    
    # Fix 2: Add validation for required fields
    validation_pattern = r'console\.log\(\'Saving campaign with data:\', campaignData\);'
    validation_replacement = '''// Validate required fields
            if (!campaignData.title) {
                showAlert('danger', 'Please enter a campaign title');
                return;
            }
            
            if (!campaignData.message) {
                showAlert('danger', 'Please enter a campaign message');
                return;
            }
            
            if (!campaignData.niche) {
                showAlert('danger', 'Please enter a niche/category');
                return;
            }
            
            console.log('Saving campaign with data:', campaignData);'''
    
    content = content.replace(validation_pattern, validation_replacement)
    
    # Fix 3: Ensure campaignDate is set when modal opens
    modal_show_pattern = r'function openCampaignModal\(date\) \{'
    modal_show_replacement = '''function openCampaignModal(date) {
            // Ensure date is valid
            if (!date) {
                const today = new Date();
                date = today.toISOString().split('T')[0];
            }'''
    
    content = content.replace(modal_show_pattern, modal_show_replacement)
    
    # Fix 4: Add event listener to ensure date is always set
    listener_pattern = r'</script>\s*</body>'
    listener_replacement = '''
        // Ensure campaign date is always set when modal is shown
        document.addEventListener('DOMContentLoaded', function() {
            const campaignModal = document.getElementById('campaignModal');
            if (campaignModal) {
                campaignModal.addEventListener('show.bs.modal', function(event) {
                    const campaignDateInput = document.getElementById('campaignDate');
                    if (campaignDateInput && !campaignDateInput.value) {
                        const today = new Date();
                        campaignDateInput.value = today.toISOString().split('T')[0];
                    }
                });
            }
        });
    </script>
</body>'''
    
    content = re.sub(listener_pattern, listener_replacement, content, flags=re.IGNORECASE)
    
    # Fix 5: Remove duplicate id="imagePreview" elements
    content = re.sub(r'<div id="imagePreview" class="mt-2"></div>\s*<div id="imagePreview" class="mt-2"></div>', 
                     '<div id="imagePreview" class="mt-2"></div>', content)
    
    # Fix 6: Fix the campaignImageUrl ID (should match what saveCampaign expects)
    content = content.replace('id="imageUrl"', 'id="campaignImageUrl"')
    
    # Write the fixed content
    with open(file_path, 'w', encoding='utf-8') as f:
        f.write(content)
    
    print(f"Fixed dashboard.html successfully!")
    print("\nChanges made:")
    print("1. Added null check for campaignDate in saveCampaign()")
    print("2. Added validation for required fields")
    print("3. Ensured date is set when modal opens")
    print("4. Added event listener for modal show event")
    print("5. Removed duplicate imagePreview elements")
    print("6. Fixed campaignImageUrl ID")

if __name__ == "__main__":
    import os
    
    # Path to dashboard.html
    dashboard_path = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\dashboard.html"
    
    if os.path.exists(dashboard_path):
        fix_dashboard_campaign_modal(dashboard_path)
    else:
        print(f"Error: Dashboard file not found at {dashboard_path}")
