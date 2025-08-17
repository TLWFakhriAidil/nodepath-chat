import subprocess
import os

os.chdir(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main')

# Get the commit before we started making changes to public_device_leads.html
# This should be around the time we first started working on the leads page

# First, let's see the recent commits
result = subprocess.run(['git', 'log', '--oneline', '-20'], capture_output=True, text=True)
print("Recent commits:")
print(result.stdout)

# Find a safe commit to revert to (before all the public device leads changes)
# Based on the commits, we'll revert to before "Fix JavaScript syntax errors"
safe_commit = "9cf9088"  # This was "Fix devices endpoint to return array format"

print(f"\nReverting public_device_leads.html to commit {safe_commit}")

# Checkout the specific file from that commit
subprocess.run(['git', 'checkout', safe_commit, 'src/views/public_device_leads.html'])

# Remove all the fix scripts we created
fix_scripts = [
    'fix_leads_parsing.py',
    'fix_load_leads.py', 
    'fix_devices_endpoint.py',
    'fix_leads_page_complete.py',
    'fix_niche_filters.py',
    'fix_leads_pagination.py',
    'add_pagination.py',
    'add_show_all.py',
    'fix_js_structure.py',
    'fix_lead_card.py',
    'fix_all_lead_issues.py',
    'fix_html_structure_final.py',
    'fix_pagination_placement.py',
    'remove_duplicate_tags.py',
    'add_import_export.py',
    'add_select_all_checkbox.py',
    'make_lead_clickable.py',
    'remove_pagination.py',
    'fix_comparison_errors.py',
    'fix_onclick_attribute.py',
    'fix_quad_equals.py',
    'fix_all_equals.py',
    'fix_template_attrs.py'
]

for script in fix_scripts:
    if os.path.exists(script):
        os.remove(script)
        print(f"Removed {script}")

print("\nReverted successfully!")
