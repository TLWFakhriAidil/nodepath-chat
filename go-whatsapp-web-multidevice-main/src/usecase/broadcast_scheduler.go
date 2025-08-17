package usecase

import (
	"math"
	"sort"
	"time"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/sirupsen/logrus"
)

// BroadcastScheduler handles scheduling conflicts for campaigns and sequences
type BroadcastScheduler struct {
	minGapMinutes int
}

// NewBroadcastScheduler creates a new scheduler
func NewBroadcastScheduler() *BroadcastScheduler {
	return &BroadcastScheduler{
		minGapMinutes: 30,
	}
}

// ScheduledBroadcast represents any scheduled broadcast (campaign or sequence)
type ScheduledBroadcast struct {
	ID           string
	Type         string    // "campaign" or "sequence"
	Title        string
	ScheduledAt  time.Time
	Status       string
	EstimatedDuration time.Duration
	MessageCount int
}

// GetScheduleConflicts checks for scheduling conflicts
func (bs *BroadcastScheduler) GetScheduleConflicts(userID string, proposedTime time.Time) ([]ScheduledBroadcast, error) {
	db := database.GetDB()
	conflicts := []ScheduledBroadcast{}
	
	// Get campaigns that might conflict
	campaignRows, err := db.Query(`
		SELECT c.id, c.title, c.scheduled_at, c.status,
		       COUNT(DISTINCT l.id) AS message_count
		FROM campaigns c
		LEFT JOIN leads l ON l.user_id = c.user_id 
		    AND (c.target_status = 'all' OR l.status = c.target_status)
		    AND (c.niche = '' OR c.niche = 'all' OR l.niche = c.niche)
		WHERE c.user_id = ?
		AND c.status IN ('pending', 'triggered', 'processing')
		AND c.scheduled_at BETWEEN ? AND ?
		GROUP BY c.id, c.title, c.scheduled_at, c.status
		ORDER BY c.scheduled_at
	`, userID, 
		proposedTime.Add(-2*time.Hour), // Check 2 hours before
		proposedTime.Add(2*time.Hour))  // Check 2 hours after
	
	if err != nil {
		return conflicts, err
	}
	defer campaignRows.Close()
	
	for campaignRows.Next() {
		var broadcast ScheduledBroadcast
		var messageCount int
		broadcast.Type = "campaign"
		
		err := campaignRows.Scan(&broadcast.ID, &broadcast.Title, 
			&broadcast.ScheduledAt, &broadcast.Status, &messageCount)
		if err != nil {
			continue
		}
		
		// Estimate duration: ~1 second per message + processing time
		broadcast.MessageCount = messageCount
		broadcast.EstimatedDuration = time.Duration(messageCount) * time.Second
		if broadcast.EstimatedDuration < 5*time.Minute {
			broadcast.EstimatedDuration = 5*time.Minute // Minimum 5 minutes
		}
		
		// Check if this conflicts with proposed time
		endTime := broadcast.ScheduledAt.Add(broadcast.EstimatedDuration)
		proposedEndTime := proposedTime.Add(5*time.Minute) // Assume minimum 5 min
		
		// Conflicts if:
		// 1. Proposed time is during this broadcast
		// 2. This broadcast is during proposed time
		// 3. Less than minGap between them
		if (proposedTime.After(broadcast.ScheduledAt) && proposedTime.Before(endTime)) ||
		   (broadcast.ScheduledAt.After(proposedTime) && broadcast.ScheduledAt.Before(proposedEndTime)) ||
		   (math.Abs(float64(broadcast.ScheduledAt.Sub(proposedTime))) < float64(bs.minGapMinutes)*float64(time.Minute)) {
			conflicts = append(conflicts, broadcast)
		}
	}
	
	// TODO: Add sequence conflict checking here
	
	return conflicts, nil
}

// AutoRescheduleCampaigns automatically reschedules campaigns to avoid conflicts
func (bs *BroadcastScheduler) AutoRescheduleCampaigns(userID string) error {
	db := database.GetDB()
	
	// Get all pending campaigns ordered by scheduled time
	rows, err := db.Query(`
		SELECT id, title, scheduled_at, 
		       (SELECT COUNT(*) FROM leads l 
		        WHERE l.user_id = c.user_id 
		        AND (c.target_status = 'all' OR l.status = c.target_status)) AS lead_count
		FROM campaigns c
		WHERE user_id = ? 
		AND status = 'pending'
		AND scheduled_at >= NOW()
		ORDER BY scheduled_at
	`, userID)
	
	if err != nil {
		return err
	}
	defer rows.Close()
	
	type campaignSchedule struct {
		ID          int
		Title       string
		ScheduledAt time.Time
		LeadCount   int
		NewTime     *time.Time
	}
	
	campaigns := []campaignSchedule{}
	for rows.Next() {
		var c campaignSchedule
		err := rows.Scan(&c.ID, &c.Title, &c.ScheduledAt, &c.LeadCount)
		if err != nil {
			continue
		}
		campaigns = append(campaigns, c)
	}
	
	// Auto-reschedule if conflicts found
	for i := 0; i < len(campaigns); i++ {
		if i == 0 {
			// First campaign keeps its time
			continue
		}
		
		prevCampaign := campaigns[i-1]
		currCampaign := campaigns[i]
		
		// Calculate estimated duration of previous campaign
		prevDuration := time.Duration(prevCampaign.LeadCount) * time.Second
		if prevDuration < 5*time.Minute {
			prevDuration = 5*time.Minute
		}
		
		// Calculate when previous campaign should finish
		prevEndTime := prevCampaign.ScheduledAt
		if prevCampaign.NewTime != nil {
			prevEndTime = *prevCampaign.NewTime
		}
		prevEndTime = prevEndTime.Add(prevDuration)
		
		// Add minimum gap
		earliestStartTime := prevEndTime.Add(time.Duration(bs.minGapMinutes) * time.Minute)
		
		// If current campaign is too close, reschedule it
		if currCampaign.ScheduledAt.Before(earliestStartTime) {
			campaigns[i].NewTime = &earliestStartTime
			
			logrus.Infof("Auto-rescheduling campaign '%s' from %s to %s to avoid conflict",
				currCampaign.Title,
				currCampaign.ScheduledAt.Format("3:04 PM"),
				earliestStartTime.Format("3:04 PM"))
			
			// Update in database
			_, err = db.Exec(`
				UPDATE campaigns 
				SET scheduled_at = ?,
				    time_schedule = ?,
				    updated_at = NOW()
				WHERE id = ?
			`, earliestStartTime, 
			   earliestStartTime.Format("15:04"),
			   currCampaign.ID)
			
			if err != nil {
				logrus.Errorf("Failed to reschedule campaign %d: %v", currCampaign.ID, err)
			}
		}
	}
	
	return nil
}

// GetBroadcastTimeline gets a timeline of scheduled broadcasts
func (bs *BroadcastScheduler) GetBroadcastTimeline(userID string, startDate, endDate time.Time) ([]ScheduledBroadcast, error) {
	timeline := []ScheduledBroadcast{}
	db := database.GetDB()
	
	// Get campaigns
	campaignRows, err := db.Query(`
		SELECT c.id, c.title, c.scheduled_at, c.status,
		       COUNT(DISTINCT l.id) AS message_count
		FROM campaigns c
		LEFT JOIN leads l ON l.user_id = c.user_id 
		WHERE c.user_id = ?
		AND c.scheduled_at BETWEEN ? AND ?
		GROUP BY c.id, c.title, c.scheduled_at, c.status
		ORDER BY c.scheduled_at
	`, userID, startDate, endDate)
	
	if err != nil {
		return timeline, err
	}
	defer campaignRows.Close()
	
	for campaignRows.Next() {
		var broadcast ScheduledBroadcast
		broadcast.Type = "campaign"
		
		err := campaignRows.Scan(&broadcast.ID, &broadcast.Title, 
			&broadcast.ScheduledAt, &broadcast.Status, &broadcast.MessageCount)
		if err != nil {
			continue
		}
		
		broadcast.EstimatedDuration = time.Duration(broadcast.MessageCount) * time.Second
		if broadcast.EstimatedDuration < 5*time.Minute {
			broadcast.EstimatedDuration = 5*time.Minute
		}
		
		timeline = append(timeline, broadcast)
	}
	
	// Sort by scheduled time
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].ScheduledAt.Before(timeline[j].ScheduledAt)
	})
	
	return timeline, nil
}
