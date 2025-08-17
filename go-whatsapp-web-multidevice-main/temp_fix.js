// Fix the displaySequences function and ensure proper syntax
        
        // Clear campaign filter
        function clearCampaignFilter() {
            document.getElementById('campaignFilterDate').value = '';
            document.getElementById('campaignRangeStart').value = '';
            document.getElementById('campaignRangeEnd').value = '';
            document.getElementById('campaignFilterInfo').style.display = 'block';
            loadCampaignSummary();
        }
        
        // Load sequences
        async function loadSequences() {
            try {
                const response = await fetch('/api/sequences', {
                    credentials: 'include'
                });
                const data = await response.json();
                
                displaySequences(data);
            } catch (error) {
                console.error('Error loading sequences:', error);
                showAlert('danger', 'Failed to load sequences');
            }
        }
        
        // Display sequences
        function displaySequences(sequences) {
            const sequencesList = document.getElementById('sequencesList');
            
            if (!sequences || sequences.length === 0) {
                sequencesList.innerHTML = `
                    <div class="col-12">
                        <div class="card">
                            <div class="card-body text-center py-5">
                                <i class="bi bi-collection display-3 text-muted mb-3"></i>
                                <h5>No Sequences Available</h5>
                                <p class="text-muted">No sequences have been created for your devices yet.</p>
                            </div>
                        </div>
                    </div>
                `;
                return;
            }
            
            sequencesList.innerHTML = sequences.map(sequence => `
                <div class="col-12 mb-3">
                    <div class="sequence-card">
                        <div class="sequence-header">
                            <h5 class="sequence-name">${sequence.name}</h5>
                            <span class="badge bg-${sequence.status === 'active' ? 'success' : 'secondary'}">
                                ${sequence.status}
                            </span>
                        </div>
                        <p class="text-muted mb-2">${sequence.description || 'No description'}</p>
                        <div class="sequence-stats">
                            <div class="sequence-stat">
                                <div class="sequence-stat-label">Total Steps</div>
                                <div class="sequence-stat-value">${sequence.total_steps || 0}</div>
                            </div>
                            <div class="sequence-stat">
                                <div class="sequence-stat-label">Active Contacts</div>
                                <div class="sequence-stat-value">${sequence.active_contacts || 0}</div>
                            </div>
                            <div class="sequence-stat">
                                <div class="sequence-stat-label">Completed</div>
                                <div class="sequence-stat-value">${sequence.completed_contacts || 0}</div>
                            </div>
                        </div>
                        <div class="sequence-actions">
                            <button class="btn btn-sm btn-outline-primary" onclick="viewSequenceDetails('${sequence.id}')">
                                <i class="bi bi-eye"></i> View Details
                            </button>
                        </div>
                    </div>
                </div>
            `).join('');
        }