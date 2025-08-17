package repository

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/models"
)

var (
	campaignRepo     CampaignRepository
	campaignRepoOnce sync.Once
)

// GetCampaignRepository returns singleton instance of CampaignRepository
func GetCampaignRepository() CampaignRepository {
	campaignRepoOnce.Do(func() {
		campaignRepo = NewCampaignRepository(database.GetDB())
	})
	return campaignRepo
}

type CampaignRepository interface {
	CreateCampaign(campaign *models.Campaign) error
	GetCampaignByDateAndNiche(scheduledDate, niche string) ([]models.Campaign, error)
	GetAllCampaigns(userID string) ([]models.Campaign, error)
	GetCampaignByID(id int) (*models.Campaign, error)
	UpdateCampaignStatus(id int, status string) error
	GetPendingCampaigns() ([]models.Campaign, error)
	// Add new methods for lead status targeting
	GetPendingCampaignsByStatus(userID string, targetStatus string) ([]models.Campaign, error)
	// Additional methods needed by the app
	GetCampaigns(userID string) ([]models.Campaign, error)
	UpdateCampaign(campaign *models.Campaign) error
	DeleteCampaign(id int) error
	GetCampaignsByUser(userID string) ([]models.Campaign, error)
	// New methods for broadcast statistics
	GetCampaignBroadcastStats(campaignID int) (shouldSend, doneSend, failedSend int, err error)
	GetUserCampaignBroadcastStats(userID string) (shouldSend, doneSend, failedSend int, err error)
	// New method for date range filtering
	GetCampaignsByUserAndDateRange(userID string, startDate string, endDate string) ([]models.Campaign, error)
}

type campaignRepository struct {
	db *sql.DB
}

func NewCampaignRepository(db *sql.DB) CampaignRepository {
	return &campaignRepository{db: db}
}

// CreateCampaign creates a new campaign
func (r *campaignRepository) CreateCampaign(campaign *models.Campaign) error {
	// Set defaults
	if campaign.MinDelaySeconds == 0 {
		campaign.MinDelaySeconds = 10
	}
	if campaign.MaxDelaySeconds == 0 {
		campaign.MaxDelaySeconds = 30
	}
	campaign.CreatedAt = time.Now()
	campaign.UpdatedAt = time.Now()
	
	query := `
		INSERT INTO campaigns(user_id, campaign_date, title, niche, target_status, message, image_url, 
		 time_schedule, min_delay_seconds, max_delay_seconds, status, ai, ` + "`limit`" + `, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	// Default target_status to 'all' if not set
	targetStatus := campaign.TargetStatus
	if targetStatus == "" {
		targetStatus = "all"
	}
	
	result, err := r.db.Exec(query, campaign.UserID, campaign.CampaignDate,
		campaign.Title, campaign.Niche, targetStatus, campaign.Message, campaign.ImageURL,
		campaign.TimeSchedule, campaign.MinDelaySeconds, campaign.MaxDelaySeconds, 
		campaign.Status, campaign.AI, campaign.Limit, campaign.CreatedAt, campaign.UpdatedAt)
		
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	campaign.ID = int(id)
	return nil
}

// GetCampaignByDateAndNiche gets campaigns by date and niche
func (r *campaignRepository) GetCampaignByDateAndNiche(scheduledDate, niche string) ([]models.Campaign, error) {
	query := `
		SELECT id, user_id, title, niche, message, image_url, 
		       campaign_date, COALESCE(time_schedule, '09:00:00') AS time_schedule, 
		       min_delay_seconds, max_delay_seconds, 
		       status, created_at, updated_at
		FROM campaigns
		WHERE campaign_date = ? AND niche = ?
	`
	
	rows, err := r.db.Query(query, scheduledDate, niche)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var campaigns []models.Campaign
	for rows.Next() {
		var c models.Campaign
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.Niche, 
			&c.Message, &c.ImageURL, &c.CampaignDate, &c.TimeSchedule, 
			&c.MinDelaySeconds, &c.MaxDelaySeconds,
			&c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	
	return campaigns, nil
}

// GetAllCampaigns gets all campaigns for a user
func (r *campaignRepository) GetAllCampaigns(userID string) ([]models.Campaign, error) {
	query := `
		SELECT id, user_id, title, niche, 
			COALESCE(target_status, 'all') AS target_status,
			message, COALESCE(image_url, '') AS image_url, campaign_date, 
			COALESCE(time_schedule, '') AS time_schedule,
			COALESCE(min_delay_seconds, 10) AS min_delay_seconds,
			COALESCE(max_delay_seconds, 30) AS max_delay_seconds,
			status, ai, COALESCE(` + "`limit`" + `, 0) AS campaign_limit, created_at, updated_at
		FROM campaigns
		WHERE user_id = ?
		ORDER BY campaign_date DESC, time_schedule DESC
	`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var campaigns []models.Campaign
	for rows.Next() {
		var c models.Campaign
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.Niche, 
			&c.TargetStatus, &c.Message, &c.ImageURL, &c.CampaignDate, 
			&c.TimeSchedule, &c.MinDelaySeconds, &c.MaxDelaySeconds,
			&c.Status, &c.AI, &c.Limit, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	
	return campaigns, nil
}

// GetCampaignByID gets a campaign by ID
func (r *campaignRepository) GetCampaignByID(id int) (*models.Campaign, error) {
	query := `
		SELECT id, user_id, title, niche, 
			COALESCE(target_status, 'all') AS target_status,
			message, COALESCE(image_url, '') AS image_url, campaign_date, 
			COALESCE(time_schedule, '') AS time_schedule,
			COALESCE(min_delay_seconds, 10) AS min_delay_seconds,
			COALESCE(max_delay_seconds, 30) AS max_delay_seconds,
			status, ai, COALESCE(` + "`limit`" + `, 0) AS campaign_limit, created_at, updated_at
		FROM campaigns
		WHERE id = ?
	`
	
	var c models.Campaign
	err := r.db.QueryRow(query, id).Scan(&c.ID, &c.UserID, &c.Title, &c.Niche, 
		&c.TargetStatus, &c.Message, &c.ImageURL, &c.CampaignDate, 
		&c.TimeSchedule, &c.MinDelaySeconds, &c.MaxDelaySeconds,
		&c.Status, &c.AI, &c.Limit, &c.CreatedAt, &c.UpdatedAt)
	
	if err != nil {
		return nil, err
	}
	
	return &c, nil
}

// UpdateCampaignStatus updates campaign status
func (r *campaignRepository) UpdateCampaignStatus(id int, status string) error {
	query := `UPDATE campaigns SET status = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, status, time.Now(), id)
	return err
}

// GetPendingCampaigns gets all campaigns with pending status
func (r *campaignRepository) GetPendingCampaigns() ([]models.Campaign, error) {
	// OPTIMIZED: Let PostgreSQL handle timezone conversions
	query := `
		SELECT id, user_id, title, niche, 
			COALESCE(target_status, 'all') AS target_status,
			message, COALESCE(image_url, '') AS image_url, campaign_date, 
			COALESCE(time_schedule, '') AS time_schedule,
			COALESCE(min_delay_seconds, 10) AS min_delay_seconds,
			COALESCE(max_delay_seconds, 30) AS max_delay_seconds,
			status, created_at, updated_at
		FROM campaigns
		WHERE status = 'pending'
		AND id NOT IN (
			-- Exclude campaigns that already have broadcast messages
			SELECT DISTINCT campaign_id 
			FROM broadcast_messages 
			WHERE campaign_id IS NOT NULL
		)
		AND (
			-- Immediate execution (no time SET)
			time_schedule IS NULL 
			OR time_schedule = ''
			-- OR scheduled time has passed - MySQL version
			OR STR_TO_DATE(CONCAT(campaign_date, ' ', time_schedule), '%Y-%m-%d %H:%i:%s') <= CONVERT_TZ(NOW(), @@session.time_zone, 'Asia/Kuala_Lumpur')
		)
		ORDER BY campaign_date, time_schedule
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		log.Printf("❌ [Campaign Repository] Error querying pending campaigns: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	var campaigns []models.Campaign
	for rows.Next() {
		var c models.Campaign
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.Niche, 
			&c.TargetStatus, &c.Message, &c.ImageURL, &c.CampaignDate, 
			&c.TimeSchedule, &c.MinDelaySeconds, &c.MaxDelaySeconds,
			&c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			log.Printf("❌ [Campaign Repository] Error scanning campaign: %v", err)
			continue
		}
		campaigns = append(campaigns, c)
	}
	
	return campaigns, nil
}

// GetPendingCampaignsByStatus gets pending campaigns filtered by target status
func (r *campaignRepository) GetPendingCampaignsByStatus(userID string, targetStatus string) ([]models.Campaign, error) {
	query := `
		SELECT id, user_id, title, niche, 
			COALESCE(target_status, 'all') AS target_status,
			message, COALESCE(image_url, '') AS image_url, campaign_date, 
			COALESCE(time_schedule, '') AS time_schedule,
			COALESCE(min_delay_seconds, 10) AS min_delay_seconds,
			COALESCE(max_delay_seconds, 30) AS max_delay_seconds,
			status, created_at, updated_at
		FROM campaigns
		WHERE user_id = ? 
		AND status = 'pending'
		AND (target_status = ? OR target_status = 'all')
		AND (
			time_schedule IS NULL 
			OR time_schedule = ''
			OR STR_TO_DATE(CONCAT(campaign_date, ' ', time_schedule), '%Y-%m-%d %H:%i:%s') <= CONVERT_TZ(NOW(), @@session.time_zone, 'Asia/Kuala_Lumpur')
		)
		ORDER BY campaign_date, time_schedule
	`
	
	rows, err := r.db.Query(query, userID, targetStatus)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var campaigns []models.Campaign
	for rows.Next() {
		var c models.Campaign
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.Niche, 
			&c.TargetStatus, &c.Message, &c.ImageURL, &c.CampaignDate, 
			&c.TimeSchedule, &c.MinDelaySeconds, &c.MaxDelaySeconds,
			&c.Status, &c.AI, &c.Limit, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	
	return campaigns, nil
}

// GetCampaigns is an alias for GetAllCampaigns
func (r *campaignRepository) GetCampaigns(userID string) ([]models.Campaign, error) {
	return r.GetAllCampaigns(userID)
}

// GetCampaignsByUser is an alias for GetAllCampaigns
func (r *campaignRepository) GetCampaignsByUser(userID string) ([]models.Campaign, error) {
	return r.GetAllCampaigns(userID)
}

// UpdateCampaign updates an existing campaign
func (r *campaignRepository) UpdateCampaign(campaign *models.Campaign) error {
	query := `
	UPDATE campaigns 
		SET title = ?, niche = ?, target_status = ?, message = ?, 
		    image_url = ?, campaign_date = ?, time_schedule = ?,
		    min_delay_seconds = ?, max_delay_seconds = ?, 
		    status = ?, ai = ?, ` + "`limit`" + ` = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`
	
	campaign.UpdatedAt = time.Now()
	
	result, err := r.db.Exec(query, 
		campaign.Title, campaign.Niche, campaign.TargetStatus, campaign.Message,
		campaign.ImageURL, campaign.CampaignDate, campaign.TimeSchedule,
		campaign.MinDelaySeconds, campaign.MaxDelaySeconds,
		campaign.Status, campaign.AI, campaign.Limit, campaign.UpdatedAt, campaign.ID, campaign.UserID)
	
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	return nil
}

// DeleteCampaign deletes a campaign by ID and its related broadcast messages
func (r *campaignRepository) DeleteCampaign(id int) error {
	// Start a transaction to ensure both deletes happen together
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// First, delete all broadcast messages for this campaign
	deleteMessagesQuery := `DELETE FROM broadcast_messages WHERE campaign_id = ?`
	_, err = tx.Exec(deleteMessagesQuery, id)
	if err != nil {
		log.Printf("Error deleting broadcast messages for campaign %d: %v", id, err)
		return err
	}
	
	// Then delete the campaign itself
	deleteCampaignQuery := `DELETE FROM campaigns WHERE id = ?`
	result, err := tx.Exec(deleteCampaignQuery, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	
	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return err
	}
	
	log.Printf("Successfully deleted campaign %d and its broadcast messages", id)
	return nil
}

// GetCampaignBroadcastStats gets broadcast statistics for a specific campaign
func (r *campaignRepository) GetCampaignBroadcastStats(campaignID int) (shouldSend, doneSend, failedSend int, err error) {
	// Get campaign details first
	campaign, err := r.GetCampaignByID(campaignID)
	if err != nil {
		return 0, 0, 0, err
	}
	
	// Get total leads that should receive the campaign based on target_status and niche
	shouldSendQuery := `
		SELECT COUNT(l.phone) 
		FROM leads l
		WHERE l.user_id = ? 
		AND l.niche LIKE CONCAT('%', ?, '%')
		AND (? = 'all' OR l.target_status = ?)
	`
	
	// Debug: Let's also get the actual leads to see what's being counted
	debugQuery := `
		SELECT l.phone, l.device_id, l.niche, l.target_status 
		FROM leads l
		WHERE l.user_id = ? 
		AND l.niche LIKE CONCAT('%', ?, '%')
		AND (? = 'all' OR l.target_status = ?)
	`
	
	rows, _ := r.db.Query(debugQuery, campaign.UserID, campaign.Niche, campaign.TargetStatus)
	if rows != nil {
		defer rows.Close()
		log.Printf("Campaign %d - Matching leads:", campaignID)
		for rows.Next() {
			var phone, deviceID, niche, targetStatus string
			rows.Scan(&phone, &deviceID, &niche, &targetStatus)
			log.Printf("  - Phone: %s, Device: %s, Niche: %s, TargetStatus: %s", phone, deviceID, niche, targetStatus)
		}
	}
	
	err = r.db.QueryRow(shouldSendQuery, campaign.UserID, campaign.Niche, campaign.TargetStatus).Scan(&shouldSend)
	if err != nil {
		return 0, 0, 0, err
	}
	
	// Debug logging
	log.Printf("Campaign %d - UserID: %s, Niche: %s, TargetStatus: %s, ShouldSend: %d", 
		campaignID, campaign.UserID, campaign.Niche, campaign.TargetStatus, shouldSend)
	
	// Get done and failed counts FROM broadcast_messages
	statsQuery := `
		SELECT COUNT(CASE WHEN status = 'success' THEN 1 END) AS done_send,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) AS failed_send
		FROM broadcast_messages
		WHERE campaign_id = ?
	`
	
	err = r.db.QueryRow(statsQuery, campaignID).Scan(&doneSend, &failedSend)
	if err != nil {
		return shouldSend, 0, 0, err
	}
	
	return shouldSend, doneSend, failedSend, nil
}

// GetUserCampaignBroadcastStats gets broadcast statistics for all campaigns of a user
func (r *campaignRepository) GetUserCampaignBroadcastStats(userID string) (shouldSend, doneSend, failedSend int, err error) {
	// Get all campaigns for the user
	campaigns, err := r.GetAllCampaigns(userID)
	if err != nil {
		return 0, 0, 0, err
	}
	
	totalShouldSend := 0
	totalDoneSend := 0
	totalFailedSend := 0
	
	// Calculate stats for each campaign and sum them up
	for _, campaign := range campaigns {
		should, done, failed, err := r.GetCampaignBroadcastStats(campaign.ID)
		if err != nil {
			// Log error but continue with other campaigns
			log.Printf("Error getting stats for campaign %d: %v", campaign.ID, err)
			continue
		}
		
		// Sum up the totals
		totalShouldSend += should
		totalDoneSend += done
		totalFailedSend += failed
		
		log.Printf("Campaign %d (%s): Should=%d, Done=%d, Failed=%d", 
			campaign.ID, campaign.Title, should, done, failed)
	}
	
	log.Printf("User %s Total Stats: Should=%d, Done=%d, Failed=%d", 
		userID, totalShouldSend, totalDoneSend, totalFailedSend)
	
	return totalShouldSend, totalDoneSend, totalFailedSend, nil
}


// GetCampaignsByUserAndDateRange gets campaigns within a specific date range
func (r *campaignRepository) GetCampaignsByUserAndDateRange(userID string, startDate string, endDate string) ([]models.Campaign, error) {
	query := `
		SELECT id, user_id, title, niche, 
			COALESCE(target_status, 'all') AS target_status,
			message, COALESCE(image_url, '') AS image_url, campaign_date, 
			COALESCE(time_schedule, '') AS time_schedule,
			COALESCE(min_delay_seconds, 10) AS min_delay_seconds,
			COALESCE(max_delay_seconds, 30) AS max_delay_seconds,
			status, ai, COALESCE(` + "`limit`" + `, 0) AS campaign_limit, created_at, updated_at
		FROM campaigns
		WHERE user_id = ?
	`
	
	args := []interface{}{userID}
	
	// Add date filters
	if startDate != "" && endDate != "" {
		query += " AND campaign_date >= ? AND campaign_date <= ?"
		args = append(args, startDate, endDate)
	} else if startDate != "" {
		query += " AND campaign_date >= ?"
		args = append(args, startDate)
	} else if endDate != "" {
		query += " AND campaign_date <= ?"
		args = append(args, endDate)
	}
	
	query += " ORDER BY campaign_date DESC, time_schedule DESC"
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var campaigns []models.Campaign
	for rows.Next() {
		var c models.Campaign
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.Niche, 
			&c.TargetStatus, &c.Message, &c.ImageURL, &c.CampaignDate, 
			&c.TimeSchedule, &c.MinDelaySeconds, &c.MaxDelaySeconds,
			&c.Status, &c.AI, &c.Limit, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	
	return campaigns, nil
}