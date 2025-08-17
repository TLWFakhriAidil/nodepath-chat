
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