// Fix for campaign summary display
// Replace the existing campaign summary cards with proper status flow

function displayCampaignSummary(summary) {
    const html = `
        <div class="row mb-4">
            <!-- Total Campaigns -->
            <div class="col-md-2">
                <div class="card text-center">
                    <div class="card-body">
                        <h3 class="text-primary">${summary.campaigns.total || 0}</h3>
                        <p class="text-muted mb-0">Total Campaigns</p>
                    </div>
                </div>
            </div>
            
            <!-- Status Flow: Pending → Triggered → Processing → Finished/Failed -->
            <div class="col-md-2">
                <div class="card text-center">
                    <div class="card-body">
                        <h3 class="text-warning">${summary.campaigns.pending || 0}</h3>
                        <p class="text-muted mb-0">Pending</p>
                        <small class="text-muted">Waiting</small>
                    </div>
                </div>
            </div>
            
            <div class="col-md-2">
                <div class="card text-center">
                    <div class="card-body">
                        <h3 class="text-info">${summary.campaigns.triggered || 0}</h3>
                        <p class="text-muted mb-0">Triggered</p>
                        <small class="text-muted">Creating Messages</small>
                    </div>
                </div>
            </div>
            
            <div class="col-md-2">
                <div class="card text-center">
                    <div class="card-body">
                        <h3 class="text-primary">${summary.campaigns.processing || 0}</h3>
                        <p class="text-muted mb-0">Processing</p>
                        <small class="text-muted">Sending</small>
                    </div>
                </div>
            </div>
            
            <div class="col-md-2">
                <div class="card text-center">
                    <div class="card-body">
                        <h3 class="text-success">${summary.campaigns.finished || 0}</h3>
                        <p class="text-muted mb-0">Finished</p>
                        <small class="text-muted">Completed</small>
                    </div>
                </div>
            </div>
            
            <div class="col-md-2">
                <div class="card text-center">
                    <div class="card-body">
                        <h3 class="text-danger">${summary.campaigns.failed || 0}</h3>
                        <p class="text-muted mb-0">Failed</p>
                        <small class="text-muted">Error</small>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Add status flow diagram -->
        <div class="row mb-4">
            <div class="col-12">
                <div class="card">
                    <div class="card-body">
                        <h6 class="card-title">Campaign Status Flow</h6>
                        <div class="d-flex justify-content-between align-items-center">
                            <div class="text-center">
                                <i class="bi bi-clock-history text-warning" style="font-size: 2rem;"></i>
                                <p class="mb-0 small">Pending</p>
                            </div>
                            <i class="bi bi-arrow-right text-muted"></i>
                            <div class="text-center">
                                <i class="bi bi-play-circle text-info" style="font-size: 2rem;"></i>
                                <p class="mb-0 small">Triggered</p>
                            </div>
                            <i class="bi bi-arrow-right text-muted"></i>
                            <div class="text-center">
                                <i class="bi bi-gear-fill text-primary" style="font-size: 2rem;"></i>
                                <p class="mb-0 small">Processing</p>
                            </div>
                            <i class="bi bi-arrow-right text-muted"></i>
                            <div class="text-center">
                                <i class="bi bi-check-circle text-success" style="font-size: 2rem;"></i>
                                <p class="mb-0 small">Finished</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Rest of the summary content... -->
    `;
    
    document.getElementById('campaignSummaryContent').innerHTML = html;
}
