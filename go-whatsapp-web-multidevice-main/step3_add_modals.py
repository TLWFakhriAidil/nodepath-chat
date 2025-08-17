# Step 3: Add modals before closing body tag
import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Find the closing container div and add modals after it
modals = '''
    <!-- Add/Edit Lead Modal -->
    <div class="modal fade" id="leadModal" tabindex="-1">
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
                                <input type="text" class="form-control" id="leadNiche" placeholder="e.g., EXSTART or EXSTART,ITADRESS">
                            </div>
                            <div class="col-md-6 mb-3">
                                <label class="form-label">Status</label>
                                <select class="form-control" id="leadStatus">
                                    <option value="prospect">Prospect</option>
                                    <option value="customer">Customer</option>
                                </select>
                            </div>
                        </div>
                        <div class="row">
                            <div class="col-md-6 mb-3">
                                <label class="form-label">Trigger</label>
                                <input type="text" class="form-control" id="leadTrigger" placeholder="e.g., EXSTART or EXSTART,ITADRESS">
                            </div>
                            <div class="col-md-6 mb-3">
                                <label class="form-label">Journey</label>
                                <input type="text" class="form-control" id="leadJourney" placeholder="e.g., Joined webinar, Downloaded guide">
                            </div>
                        </div>
                    </form>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                    <button type="button" class="btn btn-primary" onclick="saveLead()">Save Lead</button>
                </div>
            </div>
        </div>
    </div>

    <!-- Import Modal -->
    <div class="modal fade" id="importModal" tabindex="-1">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Import Leads from CSV</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <p>Upload a CSV file with the following columns:</p>
                    <ul>
                        <li><strong>name</strong> (required) - Lead's name</li>
                        <li><strong>phone</strong> (required) - Phone number in format like 60123456789</li>
                        <li><strong>niche</strong> (required) - Can be single (EXSTART) or multiple (EXSTART,ITADRESS)</li>
                        <li><strong>target_status</strong> (required) - Either "prospect" or "customer"</li>
                        <li><strong>trigger</strong> (optional) - Can be single (EXSTART) or multiple (EXSTART,ITADRESS)</li>
                    </ul>
                    <div class="alert alert-info mt-3">
                        <i class="bi bi-info-circle"></i> <strong>Note:</strong> Only these 5 columns are accepted. Any other columns will be ignored.
                    </div>
                    <div class="mt-3">
                        <input type="file" class="form-control" id="importFileInput" accept=".csv">
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                    <button type="button" class="btn btn-primary" onclick="processImport()">Import Leads</button>
                </div>
            </div>
        </div>
    </div>

    <!-- Bulk Update Modal -->
    <div class="modal fade" id="bulkUpdateModal" tabindex="-1">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Update Selected Leads</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <p class="text-muted">Update <span id="bulkUpdateCount">0</span> selected lead(s)</p>
                    <form id="bulkUpdateForm">
                        <div class="mb-3">
                            <label class="form-label">Niche (leave empty to keep current)</label>
                            <input type="text" class="form-control" id="bulkNiche" placeholder="e.g., EXSTART or EXSTART,ITADRESS">
                        </div>
                        <div class="mb-3">
                            <label class="form-label">Status (leave unchanged to keep current)</label>
                            <select class="form-control" id="bulkStatus">
                                <option value="">-- Keep Current --</option>
                                <option value="prospect">Prospect</option>
                                <option value="customer">Customer</option>
                            </select>
                        </div>
                        <div class="mb-3">
                            <label class="form-label">Trigger (leave empty to keep current)</label>
                            <input type="text" class="form-control" id="bulkTrigger" placeholder="e.g., EXSTART or EXSTART,ITADRESS">
                        </div>
                    </form>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                    <button type="button" class="btn btn-primary" onclick="saveBulkUpdate()">Update Leads</button>
                </div>
            </div>
        </div>
    </div>
'''

# Insert modals before closing container div
content = content.replace('</div>\n\n    <script src="https://cdn.jsdelivr.net/npm/bootstrap', modals + '\n</div>\n\n    <script src="https://cdn.jsdelivr.net/npm/bootstrap')

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
    f.write(content)

print("Step 3: Added all modals!")
