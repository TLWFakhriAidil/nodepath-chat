
        function updateAILeadStats(leads) {
            const stats = {
                total: leads.length,
                pending: 0,
                sent: 0,
                failed: 0
            };
            
            leads.forEach(lead => {
                switch(lead.status) {
                    case 'pending':
                        stats.pending++;
                        break;
                    case 'sent':
                        stats.sent++;
                        break;
                    case 'failed':
                        stats.failed++;
                        break;
                }
            });
            
            document.getElementById('aiTotalLeads').textContent = stats.total;
            document.getElementById('aiPendingLeads').textContent = stats.pending;
            document.getElementById('aiSentLeads').textContent = stats.sent;
            document.getElementById('aiFailedLeads').textContent = stats.failed;
        }

        function getAIStatusBadge(status) {
            const badges = {
                'pending': '<span class="badge bg-warning">Pending</span>',
                'assigned': '<span class="badge bg-info">Assigned</span>',
                'sent': '<span class="badge bg-success">Sent</span>',
                'failed': '<span class="badge bg-danger">Failed</span>'
            };
            return badges[status] || `<span class="badge bg-secondary">${status}</span>`;
        }
        function showAddAILeadModal() {
            document.getElementById('aiLeadModalTitle').textContent = 'Add AI Lead';
            document.getElementById('aiLeadForm').reset();
            document.getElementById('aiLeadId').value = '';
            $('#aiLeadModal').modal('show');
        }

        function saveAILead() {
            const leadId = document.getElementById('aiLeadId').value;
            const isEdit = !!leadId;
            
            const leadData = {
                name: document.getElementById('aiLeadName').value,
                phone: document.getElementById('aiLeadPhone').value,
                email: document.getElementById('aiLeadEmail').value,
                niche: document.getElementById('aiLeadNiche').value,
                target_status: document.getElementById('aiLeadTargetStatus').value,
                notes: document.getElementById('aiLeadNotes').value
            };
            
            if (!leadData.name || !leadData.phone) {
                showToast('Name and phone are required', 'error');
                return;
            }
            
            const url = isEdit ? `/api/leads-ai/${leadId}` : '/api/leads-ai';
            const method = isEdit ? 'PUT' : 'POST';
            
            fetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(leadData)
            })
            .then(response => response.json())
            .then(data => {
                if (data.code === 'SUCCESS') {
                    $('#aiLeadModal').modal('hide');
                    loadAILeads();
                    showToast(isEdit ? 'AI Lead updated successfully' : 'AI Lead added successfully', 'success');
                } else {
                    showToast(data.message || 'Failed to save AI lead', 'error');
                }
            })
            .catch(error => {
                console.error('Error saving AI lead:', error);
                showToast('Failed to save AI lead', 'error');
            });
        }