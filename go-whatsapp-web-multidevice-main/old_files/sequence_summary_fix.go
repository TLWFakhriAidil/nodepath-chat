// UPDATED GetSequenceSummary - Fix UI Summary Counting Logic
// Changes:
// 1. Count DISTINCT based on (sequence_stepid + recipient_phone + device_id) instead of just recipient_phone
// 2. Add total_leads count based on DISTINCT (recipient_phone + device_id)
// 3. Update Detail Sequences to also use DISTINCT (recipient_phone + device_id) for total leads

// Replace the query section in GetSequenceSummary function (around line 2250):

// Get contact statistics for this sequence FROM broadcast_messages table
var shouldSend, doneSend, failedSend, remainingSend, totalLeads int

// Build query with date filter - UPDATED to use proper DISTINCT logic
query := `
    SELECT 
        -- Count distinct combinations of (sequence_stepid + recipient_phone + device_id)
        COUNT(DISTINCT CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id)) AS total,
        COUNT(DISTINCT CASE WHEN status = 'sent' AND (error_message IS NULL OR error_message = '') 
            THEN CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id) END) AS done_send,
        COUNT(DISTINCT CASE WHEN status = 'failed' 
            THEN CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id) END) AS failed,
        COUNT(DISTINCT CASE WHEN status IN ('pending', 'queued') 
            THEN CONCAT(sequence_stepid, '|', recipient_phone, '|', device_id) END) AS remaining,
        -- NEW: Add total leads based on distinct (recipient_phone + device_id)
        COUNT(DISTINCT CONCAT(recipient_phone, '|', device_id)) AS total_leads
    FROM broadcast_messages
    WHERE sequence_id = ?`

args := []interface{}{sequence.ID}

if startDate != "" && endDate != "" {
    query += ` AND DATE(scheduled_at) BETWEEN ? AND ?`
    args = append(args, startDate, endDate)
} else if startDate != "" {
    query += ` AND DATE(scheduled_at) >= ?`
    args = append(args, startDate)
} else if endDate != "" {
    query += ` AND DATE(scheduled_at) <= ?`
    args = append(args, endDate)
}

err = db.QueryRow(query, args...).Scan(&shouldSend, &doneSend, &failedSend, &remainingSend, &totalLeads)

if err != nil {
    log.Printf("Error getting sequence stats for %s: %v", sequence.ID, err)
    shouldSend, doneSend, failedSend, remainingSend, totalLeads = 0, 0, 0, 0, 0
} else {
    // Ensure shouldSend is the sum of all statuses for consistency
    shouldSend = doneSend + failedSend + remainingSend
}

sequenceData["should_send"] = shouldSend
sequenceData["done_send"] = doneSend
sequenceData["failed_send"] = failedSend
sequenceData["remaining_send"] = remainingSend
sequenceData["total_leads"] = totalLeads  // NEW field

// Also update the summary totals section to include total_leads:

// After calculating totals from individual sequences (around line 2310):
var totalLeadsSum int
for _, seq := range sequencesWithFlows {
    if leads, ok := seq["total_leads"].(int); ok {
        totalLeadsSum += leads
    }
    // ... existing code for other totals
}

// Update the summary return (around line 2320):
summary := map[string]interface{}{
    "sequences": map[string]interface{}{
        "total": totalSequences,
        "active": activeSequences,
        "paused": pausedSequences,
        "draft": draftSequences,
    },
    "total_flows": totalFlows,
    "total_should_send": totalShouldSend,
    "total_done_send": totalDoneSend,
    "total_failed_send": totalFailedSend,
    "total_remaining_send": totalRemainingSend,
    "total_leads": totalLeadsSum,  // NEW field for total unique leads
    "contacts": map[string]interface{}{
        "total": totalContacts,
        "average_per_sequence": float64(totalContacts) / float64(max(1, totalSequences)),
    },
    "recent_sequences": sequencesWithFlows,
}