
        function editAILead(leadId) {
            fetch(`/api/leads-ai`)
                .then(response => response.json())
                .then(data => {
                    if (data.code === 'SUCCESS' && data.results) {
                        const lead = data.results.find(l => l.id === leadId);
                        if (lead) {
                            document.getElementById('aiLeadModalTitle').textContent = 'Edit AI Lead';
                            document.getElementById('aiLeadId').value = lead.id;
                            document.getElementById('aiLeadName').value = lead.name;
                            document.getElementById('aiLeadPhone').value = lead.phone;
                            document.getElementById('aiLeadEmail').value = lead.email || '';
                            document.getElementById('aiLeadNiche').value = lead.niche || '';
                            document.getElementById('aiLeadTargetStatus').value = lead.target_status;
                            document.getElementById('aiLeadNotes').value = lead.notes || '';
                            $('#aiLeadModal').modal('show');
                        }
                    }
                })
                .catch(error => {
                    console.error('Error loading lead details:', error);
                    showToast('Failed to load lead details', 'error');
                });
        }

        function deleteAILead(leadId) {
            Swal.fire({
                title: 'Delete AI Lead?',
                text: "This action cannot be undone!",
                icon: 'warning',
                showCancelButton: true,
                confirmButtonColor: '#d33',
                cancelButtonColor: '#3085d6',
                confirmButtonText: 'Yes, delete it!'
            }).then((result) => {
                if (result.isConfirmed) {
                    fetch(`/api/leads-ai/${leadId}`, {
                        method: 'DELETE'
                    })
                    .then(response => response.json())
                    .then(data => {
                        if (data.code === 'SUCCESS') {
                            loadAILeads();
                            showToast('AI Lead deleted successfully', 'success');
                        } else {
                            showToast(data.message || 'Failed to delete AI lead', 'error');
                        }
                    })
                    .catch(error => {
                        console.error('Error deleting AI lead:', error);
                        showToast('Failed to delete AI lead', 'error');
                    });
                }
            });
        }