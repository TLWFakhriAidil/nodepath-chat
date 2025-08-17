#!/usr/bin/env python3
"""
Fix the campaignImageUrl ID mismatch issue
"""

import re
from datetime import datetime

def fix_image_url_references():
    file_path = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\dashboard.html"
    
    # Read the file
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Fix 1: Find and replace remaining getElementById('imageUrl') references for campaign
    # Line 7501-7504 has the issue
    content = content.replace(
        """const campaignImageUrl = document.getElementById('imageUrl');
    if (campaignImageUrl) {
        campaignImageUrl.addEventListener('change', function() {
            validateImageUrl('imageUrl', 'imagePreview');""",
        """const campaignImageUrl = document.getElementById('campaignImageUrl');
    if (campaignImageUrl) {
        campaignImageUrl.addEventListener('change', function() {
            validateImageUrl('campaignImageUrl', 'imagePreview');"""
    )
    
    # Write the fixed content
    with open(file_path, 'w', encoding='utf-8') as f:
        f.write(content)
    
    print("Fixed campaignImageUrl references successfully!")

if __name__ == "__main__":
    fix_image_url_references()
