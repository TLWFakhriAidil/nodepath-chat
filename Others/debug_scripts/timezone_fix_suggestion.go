// Fix for campaign_trigger.go to handle timezone properly
// Add this to the ProcessCampaignTriggers function

// Option 1: Use local timezone
loc, _ := time.LoadLocation("Asia/Kuala_Lumpur")
today := time.Now().In(loc).Format("2006-01-02")

// Option 2: Get campaigns for both today and tomorrow (UTC perspective)
today := time.Now().Format("2006-01-02")
tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

// Then modify the query to check both dates:
campaigns1, _ := campaignRepo.GetCampaignsByDate(today)
campaigns2, _ := campaignRepo.GetCampaignsByDate(tomorrow)
campaigns := append(campaigns1, campaigns2...)