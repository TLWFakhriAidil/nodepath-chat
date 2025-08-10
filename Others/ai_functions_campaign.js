
        function showCreateAICampaignModal() {
            document.getElementById('aiCampaignForm').reset();
            document.getElementById('aiCampaignDate').value = new Date().toISOString().split('T')[0];
            $('#aiCampaignModal').modal('show');
        }

        function saveAICampaign() {
            const campaignData = {
                title: document.getElementById('aiCampaignTitle').value,
                niche: document.getElementById('aiCampaignNiche').value,
                target_status: document.getElementById('aiCampaignTargetStatus').value,
                message: document.getElementById('aiCampaignMessage').value,
                image_url: document.getElementById('aiCampaignImageUrl').value,
                campaign_date: document.getElementById('aiCampaignDate').value,
                time_schedule: document.getElementById('aiCampaignTime').value,
                limit: parseInt(document.getElementById('aiDeviceLimit').value),
                min_delay_seconds: parseInt(document.getElementById('aiMinDelay').value),
                max_delay_seconds: parseInt(document.getElementById('aiMaxDelay').value),
                ai: 'ai'
            };
            
            if (!campaignData.title || !campaignData.niche || !campaignData.message || 
                !campaignData.campaign_date || !campaignData.limit) {
                showToast('Please fill all required fields', 'error');
                return;
            }
            
            if (campaignData.limit <= 0) {
                showToast('Device limit must be greater than 0', 'error');
                return;
            }            
            fetch('/api/campaigns', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(campaignData)
            })
            .then(response => response.json())
            .then(data => {
                if (data.code === 'SUCCESS') {
                    $('#aiCampaignModal').modal('hide');
                    loadCampaigns();
                    showToast('AI Campaign created successfully', 'success');
                    
                    Swal.fire({
                        title: 'Campaign Created!',
                        text: 'Do you want to trigger the campaign now?',
                        icon: 'success',
                        showCancelButton: true,
                        confirmButtonText: 'Yes, trigger now',
                        cancelButtonText: 'No, trigger later'
                    }).then((result) => {
                        if (result.isConfirmed) {
                            triggerAICampaign(data.results.id);
                        }
                    });
                } else {
                    showToast(data.message || 'Failed to create AI campaign', 'error');
                }
            })
            .catch(error => {
                console.error('Error creating AI campaign:', error);
                showToast('Failed to create AI campaign', 'error');
            });
        }

        function triggerAICampaign(campaignId) {
            showToast('Triggering AI campaign...', 'info');
            
            fetch(`/api/campaigns-ai/${campaignId}/trigger`, {
                method: 'POST'
            })
            .then(response => response.json())
            .then(data => {
                if (data.code === 'SUCCESS') {
                    showToast('AI Campaign triggered successfully! Processing leads...', 'success');
                    loadCampaigns();
                } else {
                    showToast(data.message || 'Failed to trigger AI campaign', 'error');
                }
            })
            .catch(error => {
                console.error('Error triggering AI campaign:', error);
                showToast('Failed to trigger AI campaign', 'error');
            });
        }