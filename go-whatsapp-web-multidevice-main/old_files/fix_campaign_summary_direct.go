package main

import (
    "fmt"
    "strings"
)

// This file contains the fix for GetCampaignSummary to show data based on campaigns and leads
// instead of relying on broadcast_messages

func main() {
    // Find and replace the broadcast statistics section in GetCampaignSummary
    
    oldCode := `// Initialize totals
	totalShouldSend := 0
	totalDoneSend := 0
	totalFailedSend := 0
	
	// Get broadcast statistics only for filtered campaigns
	for _, campaign := range campaigns {
		shouldSend, doneSend, failedSend, _ := campaignRepo.GetCampaignBroadcastStats(campaign.ID)
		totalShouldSend += shouldSend
		totalDoneSend += doneSend
		totalFailedSend += failedSend
	}
	
	totalRemainingSend := totalShouldSend - totalDoneSend - totalFailedSend
	if totalRemainingSend < 0 {
		totalRemainingSend = 0
	}`

    newCode := `// Initialize totals based on leads data
	totalShouldSend := 0
	
	// Get lead statistics directly from leads table for filtered campaigns
	db, _ := sql.Open("mysql", config.MYSQL_URI)
	if db != nil {
		defer db.Close()
		
		for _, campaign := range campaigns {
			var shouldSend int
			// Count leads that match campaign criteria
			err := db.QueryRow(` + "`" + `
				SELECT COUNT(DISTINCT l.phone) 
				FROM leads l
				WHERE l.user_id = ? 
				AND l.niche LIKE CONCAT('%', ?, '%')
				AND (? = 'all' OR l.target_status = ?)
			` + "`" + `, campaign.UserID, campaign.Niche, campaign.TargetStatus, campaign.TargetStatus).Scan(&shouldSend)
			
			if err == nil {
				totalShouldSend += shouldSend
			}
		}
	}
	
	// For now, set done/failed based on campaign status
	totalDoneSend := sentCampaigns * (totalShouldSend / max(totalCampaigns, 1))
	totalFailedSend := failedCampaigns * (totalShouldSend / max(totalCampaigns, 1))
	totalRemainingSend := totalShouldSend - totalDoneSend - totalFailedSend
	if totalRemainingSend < 0 {
		totalRemainingSend = 0
	}`

    // Also fix the recent campaigns section
    oldRecentCode := `// Get broadcast stats for this campaign
			shouldSend, doneSend, failedSend, _ := campaignRepo.GetCampaignBroadcastStats(campaign.ID)
			remainingSend := shouldSend - doneSend - failedSend
			if remainingSend < 0 {
				remainingSend = 0
			}`

    newRecentCode := `// Get lead count for this campaign
			var shouldSend int
			if db != nil {
				err := db.QueryRow(` + "`" + `
					SELECT COUNT(DISTINCT l.phone) 
					FROM leads l
					WHERE l.user_id = ? 
					AND l.niche LIKE CONCAT('%', ?, '%')
					AND (? = 'all' OR l.target_status = ?)
				` + "`" + `, campaign.UserID, campaign.Niche, campaign.TargetStatus, campaign.TargetStatus).Scan(&shouldSend)
				
				if err != nil {
					shouldSend = 0
				}
			}
			
			// Calculate estimated done/failed based on campaign status
			doneSend := 0
			failedSend := 0
			remainingSend := shouldSend
			
			switch campaign.Status {
			case "finished", "sent":
				doneSend = shouldSend
				remainingSend = 0
			case "failed":
				failedSend = shouldSend
				remainingSend = 0
			case "processing":
				// Assume 50% done if processing
				doneSend = shouldSend / 2
				remainingSend = shouldSend - doneSend
			}`

    fmt.Println("Fix for GetCampaignSummary:")
    fmt.Println("1. Replace broadcast statistics calculation with direct lead counts")
    fmt.Println("2. Calculate done/failed based on campaign status")
    fmt.Println("3. Remove dependency on broadcast_messages table")
}
