// Fix for campaign creation/editing issues in dashboard.html

// The main issue is that campaignDate might be null when saveCampaign is called
// This can happen if the modal is opened without going through openCampaignModal

// Solution 1: Add null check and default value in saveCampaign function
function saveCampaignFixed() {
    const campaignId = document.getElementById('campaignId').value;
    let campaignDate = document.getElementById('campaignDate').value;
    
    // If campaignDate is null or empty, use today's date
    if (!campaignDate) {
        const today = new Date();
        campaignDate = today.toISOString().split('T')[0]; // Format: YYYY-MM-DD
        console.warn('Campaign date was empty, using today:', campaignDate);
    }
    
    // Auto-set current time if not provided
    let scheduledTime = document.getElementById('campaignTime').value;
    if (!scheduledTime) {
        const now = new Date();
        scheduledTime = now.toTimeString().slice(0, 5); // HH:MM format
    }
    
    const campaignData = {
        campaign_date: campaignDate,
        title: document.getElementById('campaignTitle').value || '',
        niche: document.getElementById('campaignNiche').value || '',
        target_status: document.getElementById('campaignTargetStatus').value || 'prospect',
        message: document.getElementById('campaignMessage').value || '',
        image_url: document.getElementById('campaignImageUrl').value || '',
        time_schedule: scheduledTime,
        min_delay_seconds: parseInt(document.getElementById('campaignMinDelay').value) || 10,
        max_delay_seconds: parseInt(document.getElementById('campaignMaxDelay').value) || 30
    };
    
    // Validate required fields
    if (!campaignData.title) {
        alert('Please enter a campaign title');
        return;
    }
    
    if (!campaignData.message) {
        alert('Please enter a campaign message');
        return;
    }
    
    console.log('Saving campaign with data:', campaignData);
    
    const url = campaignId ? `/api/campaigns/${campaignId}` : '/api/campaigns';
    const method = campaignId ? 'PUT' : 'POST';
    
    fetch(url, {
        method: method,
        headers: {
            'Content-Type': 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify(campaignData)
    })
    .then(response => response.json())
    .then(data => {
        if (data.code === 'SUCCESS') {
            bootstrap.Modal.getInstance(document.getElementById('campaignModal')).hide();
            loadCampaigns();
            showAlert('success', campaignId ? 'Campaign updated successfully' : 'Campaign created successfully');
        } else {
            showAlert('danger', data.message || 'Failed to save campaign');
        }
    })
    .catch(error => {
        console.error('Error saving campaign:', error);
        showAlert('danger', 'Error saving campaign');
    });
}

// Solution 2: Ensure campaignDate is set when modal is shown
function ensureCampaignDateSet() {
    const campaignDateInput = document.getElementById('campaignDate');
    if (campaignDateInput && !campaignDateInput.value) {
        const today = new Date();
        campaignDateInput.value = today.toISOString().split('T')[0];
    }
}

// Solution 3: Add event listener to set date when modal is shown
document.addEventListener('DOMContentLoaded', function() {
    const campaignModal = document.getElementById('campaignModal');
    if (campaignModal) {
        campaignModal.addEventListener('show.bs.modal', function(event) {
            // Ensure campaign date is set
            ensureCampaignDateSet();
        });
    }
});

// Instructions to apply the fix:
// 1. Replace the saveCampaign function in dashboard.html with saveCampaignFixed
// 2. Add the ensureCampaignDateSet function to dashboard.html
// 3. Add the event listener code to ensure date is always set when modal opens
