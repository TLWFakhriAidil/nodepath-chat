package whatsapp

import (
	"time"
)

// GetMalaysiaTime converts a timestamp to Malaysia timezone
func GetMalaysiaTime(timestamp int64) time.Time {
	// Load Malaysia timezone (UTC+8)
	loc, err := time.LoadLocation("Asia/Kuala_Lumpur")
	if err != nil {
		// Fallback to fixed UTC+8 if timezone data not available
		loc = time.FixedZone("MYT", 8*60*60)
	}
	
	// Convert Unix timestamp to time with Malaysia timezone
	return time.Unix(timestamp, 0).In(loc)
}

// FormatMessageTimeMalaysia formats timestamp to readable time in Malaysia timezone
func FormatMessageTimeMalaysia(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}
	
	malaysiaTime := GetMalaysiaTime(timestamp)
	now := time.Now().In(malaysiaTime.Location())
	
	// Today - show time only
	if malaysiaTime.Day() == now.Day() && malaysiaTime.Month() == now.Month() && malaysiaTime.Year() == now.Year() {
		return malaysiaTime.Format("15:04") // 24-hour format
	}
	
	// Yesterday
	yesterday := now.AddDate(0, 0, -1)
	if malaysiaTime.Day() == yesterday.Day() && malaysiaTime.Month() == yesterday.Month() && malaysiaTime.Year() == yesterday.Year() {
		return "Yesterday"
	}
	
	// This week - show day name
	if now.Sub(malaysiaTime) < 7*24*time.Hour {
		return malaysiaTime.Format("Monday")
	}
	
	// This year - show date without year
	if malaysiaTime.Year() == now.Year() {
		return malaysiaTime.Format("Jan 2")
	}
	
	// Older - show full date
	return malaysiaTime.Format("02/01/2006")
}