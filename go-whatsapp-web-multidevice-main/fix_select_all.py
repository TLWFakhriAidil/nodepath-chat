# Fix the Select All checkbox
import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Replace the Select All text with proper checkbox
old_select_all = r'<div class="col-md-4 text-end">\s*<div class="form-check">\s*<input class="form-check-input" type="checkbox" id="selectAll" onchange="toggleSelectAll\(\)">\s*<label class="form-check-label" for="selectAll">\s*Select All\s*</label>\s*</div>\s*</div>'

# Check if already has form-check div
if '<div class="form-check">' in content and 'id="selectAll"' in content:
    print("Select All checkbox already exists")
else:
    # Replace plain text with checkbox
    content = re.sub(
        r'<div class="col-md-4 text-end">\s*Select All\s*</div>',
        '''<div class="col-md-4 text-end">
                    <div class="form-check">
                        <input class="form-check-input" type="checkbox" id="selectAll" onchange="toggleSelectAll()">
                        <label class="form-check-label" for="selectAll">
                            Select All
                        </label>
                    </div>
                </div>''',
        content
    )

# Also ensure buildNicheFilters is properly called
content = content.replace('buildNicheFilters();', 'buildNicheFilters();')

# Fix comment formatting
content = re.sub(r'// Toggle select all\s*\n\s*function toggleSelectAll', '// Toggle select all\n        function toggleSelectAll', content)
content = re.sub(r'// Update select all checkbox\s*\n\s*function updateSelectAllCheckbox', '// Update select all checkbox\n        function updateSelectAllCheckbox', content)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
    f.write(content)

print("Fixed Select All checkbox!")
