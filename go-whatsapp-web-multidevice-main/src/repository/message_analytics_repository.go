package repository

import (
	"database/sql"
	"fmt"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/google/uuid"
)

// MessageAnalytics represents a message record
type MessageAnalytics struct {
	ID        string
	UserID    string
	DeviceID  string
	MessageID string
	JID       string
	Content   string
	IsFromMe  bool
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// MessageAnalyticsRepository handles message analytics persistence
type MessageAnalyticsRepository struct {
	db *sql.DB
}

// NewMessageAnalyticsRepository creates a new repository
func NewMessageAnalyticsRepository() *MessageAnalyticsRepository {
	return &MessageAnalyticsRepository{
		db: database.GetDB(),
	}
}

// RecordMessage records a new message
func (r *MessageAnalyticsRepository) RecordMessage(userID, deviceID, messageID, jid, content string, isFromMe bool, status string) error {
	id := uuid.New().String()
	query := `
		INSERT INTO message_analytics(id, user_id, device_id, message_id, jid, content, is_from_me, ` + "`status`" + `)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (message_id) 
		DO UPDATE SET status = ?, updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.Exec(query, id, userID, deviceID, messageID, jid, content, isFromMe, status)
	if err != nil {
		return fmt.Errorf("failed to record message: %w", err)
	}
	
	return nil
}

// UpdateMessageStatus updates the status of a message
func (r *MessageAnalyticsRepository) UpdateMessageStatus(messageID, status string) error {
	query := `
		UPDATE message_analytics SET ` + "`status`" + ` = ?, updated_at = CURRENT_TIMESTAMP
		WHERE message_id = ?
	`
	_, err := r.db.Exec(query, messageID, status)
	if err != nil {
		return fmt.Errorf("failed to update message status: %w", err)
	}
	
	return nil
}

// GetUserAnalytics gets analytics for a user
func (r *MessageAnalyticsRepository) GetUserAnalytics(userID string, startDate, endDate time.Time, deviceID string) (map[string]interface{}, error) {
	// Base query for metrics
	metricsQuery := `
		SELECT COUNT(CASE WHEN is_from_me = true THEN 1 END) as leads_sent,
			COUNT(CASE WHEN is_from_me = true AND status IN ('delivered', 'read') THEN 1 END) as leads_delivered,
			COUNT(CASE WHEN is_from_me = true AND ` + "`status`" + ` = 'read' THEN 1 END) as leads_read,
			COUNT(CASE WHEN is_from_me = false THEN 1 END) as leads_replied,
			COUNT(DISTINCT CASE WHEN device_id IS NOT NULL THEN device_id END) as active_devices
		FROM message_analytics
		WHERE user_id = ? AND created_at BETWEEN ? AND ?
	`
	
	args := []interface{}{userID, startDate, endDate}
	if deviceID != "" && deviceID != "all" {
		metricsQuery += " AND device_id = ?"
		args = append(args, deviceID)
	}
	
	var leadsSent, leadsDelivered, leadsRead, leadsReplied, activeDevices int
	err := r.db.QueryRow(metricsQuery, args...).Scan(
		&leadsSent, &leadsDelivered, &leadsRead, &leadsReplied, &activeDevices,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	
	// Calculate derived metrics
	leadsNotReceived := leadsSent - leadsDelivered
	leadsNotRead := leadsDelivered - leadsRead
	
	// Get device count
	var totalDevices int
	deviceQuery := "SELECT COUNT(*) `from` user_devices WHERE user_id = ?"
	r.db.QueryRow(deviceQuery, userID).Scan(&totalDevices)
	inactiveDevices := totalDevices - activeDevices
	
	// Get daily stats
	dailyQuery := `
		SELECT DATE(created_at) as date,
			COUNT(CASE WHEN is_from_me = true THEN 1 END) as sent,
			COUNT(CASE WHEN is_from_me = true AND status IN ('delivered', 'read') THEN 1 END) as delivered,
			COUNT(CASE WHEN is_from_me = true AND ` + "`status`" + ` = 'read' THEN 1 END) as read,
			COUNT(CASE WHEN is_from_me = false THEN 1 END) as replied
		FROM message_analytics
		WHERE user_id = ? AND created_at BETWEEN ? AND ?
	`
	
	dailyArgs := []interface{}{userID, startDate, endDate}
	if deviceID != "" && deviceID != "all" {
		dailyQuery += " AND device_id = ?"
		dailyArgs = append(dailyArgs, deviceID)
	}
	dailyQuery += " GROUP BY DATE(created_at) ORDER BY date"
	
	rows, err := r.db.Query(dailyQuery, dailyArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stats: %w", err)
	}
	defer rows.Close()
	
	var dailyStats []map[string]interface{}
	for rows.Next() {
		var date time.Time
		var sent, delivered, read, replied int
		
		err := rows.Scan(&date, &sent, &delivered, &read, &replied)
		if err != nil {
			continue
		}
		
		dailyStats = append(dailyStats, map[string]interface{}{
			"date":      date.Format("Jan 2"),
			"sent":      sent,
			"delivered": delivered,
			"read":      read,
			"replied":   replied,
		})
	}
	
	// Fill in missing dates with zeros
	fullDailyStats := fillMissingDates(dailyStats, startDate, endDate)
	
	return map[string]interface{}{
		"metrics": map[string]interface{}{
			"activeDevices":     activeDevices,
			"inactiveDevices":   inactiveDevices,
			"leadsSent":         leadsSent,
			"leadsReceived":     leadsDelivered,
			"leadsNotReceived":  leadsNotReceived,
			"leadsRead":         leadsRead,
			"leadsNotRead":      leadsNotRead,
			"leadsReplied":      leadsReplied,
		},
		"daily": fullDailyStats,
	}, nil
}

// fillMissingDates ensures all dates in range have data
func fillMissingDates(dailyStats []map[string]interface{}, startDate, endDate time.Time) []map[string]interface{} {
	dateMap := make(map[string]map[string]interface{})
	
	// Create map of existing data
	for _, stat := range dailyStats {
		dateStr := stat["date"].(string)
		dateMap[dateStr] = stat
	}
	
	// Fill in missing dates
	var result []map[string]interface{}
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("Jan 2")
		if stat, exists := dateMap[dateStr]; exists {
			result = append(result, stat)
		} else {
			result = append(result, map[string]interface{}{
				"date":      dateStr,
				"sent":      0,
				"delivered": 0,
				"read":      0,
				"replied":   0,
			})
		}
	}
	
	return result
}

// GetDeviceAnalytics gets analytics for a specific device
func (r *MessageAnalyticsRepository) GetDeviceAnalytics(deviceID string, days int) (map[string]interface{}, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)
	
	query := `

		SELECT COUNT(CASE WHEN is_from_me = true THEN 1 END) AS messages_sent,
			COUNT(CASE WHEN is_from_me = false THEN 1 END) AS messages_received,
			COUNT(DISTINCT jid) AS unique_contacts
		FROM message_analytics
		WHERE device_id = ? AND created_at BETWEEN ? AND ?
	`
	
	var sent, received, contacts int
	err := r.db.QueryRow(query, deviceID, startDate, endDate).Scan(&sent, &received, &contacts)
	if err != nil {
		return nil, fmt.Errorf("failed to get device analytics: %w", err)
	}
	
	return map[string]interface{}{
		"messagesSent":     sent,
		"messagesReceived": received,
		"uniqueContacts":   contacts,
		"period":           days,
	}, nil
}