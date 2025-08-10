// Update for displayCampaignSummary function

function displayCampaignSummary(summary) {
    // Ensure broadcast_stats exists
    summary.broadcast_stats = summary.broadcast_stats || {
        total_should_send: 0,
        total_done_send: 0,
        total_failed_send: 0,
        total_remaining_send: 0
    };
    
    const html = `
        <div class="row mb-4">
            <!-- Total Campaigns -->
            <div class="col-6 col-md-4 col-lg-2 mb-3">
                <div class="card text-center">
                    <div class="card-body">
                        <h3 class="text-primary">${summary.campaigns.total || 0}</h3>
                        <p class="text-muted mb-0">Total Campaigns</p>
                    </div>
                </div>
            </div>
            
            <!-- Total Contacts Should Send -->
            <div class="col-6 col-md-4 col-lg-2 mb-3">
                <div class="card text-center">
                    <div class="card-body">
                        <h3 class="text-primary">${summary.broadcast_stats.total_should_send || 0}</h3>
                        <p class="text-muted mb-0">Total Contacts Should Send</p>
                    </div>
                </div>
            </div>
            
            <!-- Contacts Done Send Message -->
            <div class="col-6 col-md-4 col-lg-2 mb-3">
                <div class="card text-center">
                    <div class="card-body">
                        <h3 class="text-success">${summary.broadcast_stats.total_done_send || 0}</h3>
                        <p class="text-muted mb-0">Contacts Done Send Message</p>
                    </div>
                </div>
            </div>
            
            <!-- Contacts Failed Send Message -->
            <div class="col-6 col-md-4 col-lg-2 mb-3">
                <div class="card text-center">
                    <div class="card-body">
                        <h3 class="text-danger">${summary.broadcast_stats.total_failed_send || 0}</h3>
                        <p class="text-muted mb-0">Contacts Failed Send Message</p>
                    </div>
                </div>
            </div>
            
            <!-- Contacts Remaining Send Message -->
            <div class="col-6 col-md-4 col-lg-2 mb-3">
                <div class="card text-center">
                    <div class="card-body">
                        <h3 class="text-warning">${summary.broadcast_stats.total_remaining_send || 0}</h3>
                        <p class="text-muted mb-0">Contacts Remaining Send Message</p>
                    </div>
                </div>
            </div>
            
            <!-- Status Cards -->
            <div class="col-6 col-md-4 col-lg-2 mb-3">
                <div class="card text-center">
                    <div class="card-body">
                        <h3 class="text-info">${summary.campaigns.pending || 0}</h3>
                        <p class="text-muted mb-0">Pending</p>
                    </div>
                </div>
            </div>
        </div>
        
        <div class="row">
            <div class="col-12">
                <div class="card">
                    <div class="card-header">
                        <h5 class="mb-0">Recent Campaigns</h5>
                    </div>
                    <div class="card-body">
                        ${summary.recent_campaigns.length > 0 ? `
                            <div class="table-responsive">
                                <table class="table">
                                    <thead>
                                        <tr>
                                            <th>Title</th>
                                            <th>Date</th>
                                            <th>Time</th>
                                            <th>Niche</th>
                                            <th>Target Status</th>
                                            <th>Status</th>
                                            <th>Contacts Should Send</th>
                                            <th>Contacts Done Send Message</th>
                                            <th>Contacts Failed Send Message</th>
                                            <th>Contacts Remaining Send Message</th>
                                            <th>Actions</th>
                                            <th>Device Report</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        ${summary.recent_campaigns.map(campaign => `
                                            <tr>
                                                <td>${campaign.title}</td>
                                                <td>${campaign.campaign_date ? formatCampaignDate(campaign.campaign_date) : '-'}</td>
                                                <td>${campaign.time_schedule || '-'}</td>
                                                <td>${campaign.niche || '-'}</td>
                                                <td>
                                                    <span class="badge bg-${campaign.target_status === 'all' ? 'info' : campaign.target_status === 'prospect' ? 'primary' : 'success'}">
                                                        ${campaign.target_status || 'all'}
                                                    </span>
                                                </td>
                                                <td>
                                                    <span class="badge bg-${getStatusColor(campaign.status)}">
                                                        ${campaign.status}
                                                    </span>
                                                </td>
                                                <td>${campaign.should_send || 0}</td>
                                                <td>${campaign.done_send || 0}</td>
                                                <td>${campaign.failed_send || 0}</td>
                                                <td>${campaign.remaining_send || 0}</td>
                                                <td>
                                                    <button class="btn btn-sm btn-outline-primary" onclick="previewCampaignMessage(${JSON.stringify(campaign).replace(/"/g, '&quot;')})" title="Preview Message">
                                                        <i class="bi bi-eye"></i>
                                                    </button>
                                                </td>
                                                <td>
                                                    <button class="btn btn-sm btn-outline-success" onclick="showCampaignDeviceReport(${JSON.stringify(campaign).replace(/"/g, '&quot;')})" title="Device Report">
                                                        <i class="bi bi-phone"></i> Report
                                                    </button>
                                                </td>
                                            </tr>
                                        `).join('')}
                                    </tbody>
                                </table>
                            </div>
                        ` : '<p class="text-muted text-center">No campaigns yet</p>'}
                    </div>
                </div>
            </div>
        </div>
    `;
    document.getElementById('campaignSummaryContent').innerHTML = html;
}
