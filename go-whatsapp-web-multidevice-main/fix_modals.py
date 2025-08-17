# Fix 1: Update the Add Lead Modal to match the original
import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Find and replace the Add Lead Modal
old_modal = r'<div class="modal fade" id="leadModal" tabindex="-1">[\s\S]*?</div>\s*</div>\s*</div>'

new_modal = '''<div class="modal fade" id="leadModal" tabindex="-1">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title" id="leadModalTitle">Add New Lead</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <form id="leadForm">
                        <input type="hidden" id="leadId">
                        <div class="row">
                            <div class="col-md-6 mb-3">
                                <label class="form-label">Name <span class="text-danger">*</span></label>
                                <input type="text" class="form-control" id="leadName" required>
                            </div>
                            <div class="col-md-6 mb-3">
                                <label class="form-label">Phone Number <span class="text-danger">*</span></label>
                                <input type="text" class="form-control" id="leadPhone" placeholder="60123456789" required>
                            </div>
                        </div>
                        <div class="row">
                            <div class="col-md-6 mb-3">
                                <label class="form-label">Niche</label>
                                <input type="text" class="form-control" id="leadNiche" placeholder="EXSTART or EXSTART,ITADRESS">
                            </div>
                            <div class="col-md-6 mb-3">
                                <label class="form-label">Status</label>
                                <select class="form-control" id="leadStatus">
                                    <option value="prospect">Prospect</option>
                                    <option value="customer">Customer</option>
                                </select>
                            </div>
                        </div>
                        <div class="mb-3">
                            <label class="form-label">Sequence Triggers</label>
                            <input type="text" class="form-control" id="leadTrigger" placeholder="fitness_start,crypto_welcome (comma-separated)">
                            <small class="text-muted">Enter sequence triggers to auto-enroll this lead</small>
                        </div>
                        <div class="mb-3">
                            <label class="form-label">Additional Note</label>
                            <textarea class="form-control" id="leadJourney" rows="3" placeholder="Enter additional notes about this lead..."></textarea>
                        </div>
                    </form>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                    <button type="button" class="btn btn-primary" onclick="saveLead()">Save Lead</button>
                </div>
            </div>
        </div>
    </div>'''

content = re.sub(old_modal, new_modal, content, flags=re.DOTALL)

# Fix 2: Update Bulk Update Modal
old_bulk_modal = r'<div class="modal fade" id="bulkUpdateModal" tabindex="-1">[\s\S]*?</div>\s*</div>\s*</div>'

new_bulk_modal = '''<div class="modal fade" id="bulkUpdateModal" tabindex="-1">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Bulk Update Selected Leads</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <p class="text-muted">Update <span id="bulkUpdateCount">0</span> selected leads. Leave fields empty to keep existing values.</p>
                    <form id="bulkUpdateForm">
                        <div class="mb-3">
                            <label class="form-label">Customer Name</label>
                            <input type="text" class="form-control" id="bulkName" placeholder="Customer name">
                            <small class="text-muted">Leave empty to keep existing values</small>
                        </div>
                        <div class="mb-3">
                            <label class="form-label">Niche</label>
                            <input type="text" class="form-control" id="bulkNiche" placeholder="EXSTART or EXSTART,ITADRESS">
                            <small class="text-muted">Leave empty to keep existing values</small>
                        </div>
                        <div class="mb-3">
                            <label class="form-label">Target Status</label>
                            <select class="form-control" id="bulkStatus">
                                <option value="">Keep existing</option>
                                <option value="prospect">Prospect</option>
                                <option value="customer">Customer</option>
                            </select>
                        </div>
                        <div class="mb-3">
                            <label class="form-label">Sequence Triggers</label>
                            <input type="text" class="form-control" id="bulkTrigger" placeholder="fitness_start,crypto_welcome (comma-separated)">
                            <small class="text-muted">Leave empty to keep existing values</small>
                        </div>
                    </form>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                    <button type="button" class="btn btn-primary" onclick="saveBulkUpdate()">Update Selected</button>
                </div>
            </div>
        </div>
    </div>'''

content = re.sub(old_bulk_modal, new_bulk_modal, content, flags=re.DOTALL)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
    f.write(content)

print("Fixed modals!")
