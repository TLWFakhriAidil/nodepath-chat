# Step 2: Add Select All checkbox and bulk action buttons
import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Find the col-md-4 text-end div and replace with checkbox
old_div = r'<div class="col-md-4 text-end">\s*<!-- Placeholder for bulk actions -->\s*</div>'
new_div = '''<div class="col-md-4 text-end">
                    <div class="form-check">
                        <input class="form-check-input" type="checkbox" id="selectAll" onchange="toggleSelectAll()">
                        <label class="form-check-label" for="selectAll">
                            Select All
                        </label>
                    </div>
                </div>'''

content = re.sub(old_div, new_div, content)

# Add bulk actions section after filters
filters_end = '<!-- Leads List -->'
bulk_actions = '''<!-- Bulk Actions -->
        <div class="bulk-actions" id="bulkActions" style="display: none;">
            <span class="selected-count">
                <span id="selectedCount">0</span> selected
            </span>
            <button class="btn btn-sm btn-danger" onclick="bulkDelete()">
                <i class="bi bi-trash"></i> Delete Selected
            </button>
            <button class="btn btn-sm btn-warning ms-2" onclick="bulkUpdate()">
                <i class="bi bi-pencil"></i> Update Selected
            </button>
        </div>

        <!-- Leads List -->'''

content = content.replace(filters_end, bulk_actions)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
    f.write(content)

print("Step 2: Added Select All checkbox and bulk actions!")
