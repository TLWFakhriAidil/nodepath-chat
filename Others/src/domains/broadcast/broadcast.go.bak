package broadcast

import "time"

// BroadcastMessage represents a message to be broadcast
type BroadcastMessage struct {
	ID           string
	DeviceID     string
	RecipientJID string
	Message      string
	ImageURL     string
	CampaignID   *int
	SequenceID   *string
	GroupID      string
	GroupOrder   int
	RetryCount   int
	CreatedAt    time.Time
	// Delay settings from campaign/sequence
	MinDelay     int
	MaxDelay     int
}
