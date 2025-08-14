package repository

import (
	"database/sql"
	"strings"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/models"
	"github.com/sirupsen/logrus"
)

type LeadAIRepository interface {
	CreateLeadAI(lead *models.LeadAI) error
	GetLeadAIByID(id int) (*models.LeadAI, error)
	GetLeadAIByUser(userID string) ([]models.LeadAI, error)
	GetPendingLeadAI(userID string) ([]models.LeadAI, error)
	GetLeadAIByNiche(userID, niche string) ([]models.LeadAI, error)
	GetLeadAIByNicheAndStatus(userID, niche, targetStatus string) ([]models.LeadAI, error)
	AssignDevice(leadID int, deviceID string) error
	UpdateStatus(leadID int, status string) error
	UpdateLeadAI(id int, lead *models.LeadAI) error
	DeleteLeadAI(id int) error
	GetLeadAICountByDevice(campaignID int, deviceID string) (int, error)
	GetCampaignProgress(campaignID int) ([]models.AICampaignProgress, error)
	UpdateCampaignProgress(progress *models.AICampaignProgress) error
}

type leadAIRepository struct {
	db  *sql.DB
}

var leadAIRepo *leadAIRepository
func GetLeadAIRepository() LeadAIRepository {
	if leadAIRepo == nil {
		leadAIRepo = &leadAIRepository{
			db: database.GetDB(),
		}
	}
	return leadAIRepo
}

func (r *leadAIRepository) CreateLeadAI(lead *models.LeadAI) error {
	query := `

		INSERT INTO leads_ai(user_id, name, phone, email, niche, source, target_status, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`
	
	err := r.db.QueryRow(
		query,
		lead.UserID,
		lead.Name,
		lead.Phone,
		lead.Email,
		lead.Niche,
		lead.Source,
		lead.TargetStatus,
		lead.Notes,
	).Scan(&lead.ID, &lead.CreatedAt, &lead.UpdatedAt)
	
	return err
}

func (r *leadAIRepository) GetLeadAIByID(id int) (*models.LeadAI, error) {
	var lead models.LeadAI
	query := `

		SELECT id, user_id, device_id, name, phone, email, niche, source, 
		       status, target_status, notes, assigned_at, sent_at, created_at, updated_at
		FROM leads_ai
		WHERE id = ?`
	
	err := r.db.QueryRow(query, id).Scan(
		&lead.ID,
		&lead.UserID,
		&lead.DeviceID,
		&lead.Name,
		&lead.Phone,
		&lead.Email,
		&lead.Niche,
		&lead.Source,
		&lead.Status,
		&lead.TargetStatus,
		&lead.Notes,
		&lead.AssignedAt,
		&lead.SentAt,
		&lead.CreatedAt,
		&lead.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &lead, nil
}
func (r *leadAIRepository) GetLeadAIByUser(userID string) ([]models.LeadAI, error) {
	query := `

		SELECT id, user_id, device_id, name, phone, email, niche, source, 
		       status, target_status, notes, assigned_at, sent_at, created_at, updated_at
		FROM leads_ai
		WHERE user_id = ?
		ORDER BY created_at DESC`
	
	return r.getLeadAIList(query, userID)
}

func (r *leadAIRepository) GetPendingLeadAI(userID string) ([]models.LeadAI, error) {
	query := `

		SELECT id, user_id, device_id, name, phone, email, niche, source, 
		       status, target_status, notes, assigned_at, sent_at, created_at, updated_at
		FROM leads_ai
		WHERE user_id = ? AND status = 'pending'
		ORDER BY created_at ASC`
	
	return r.getLeadAIList(query, userID)
}

func (r *leadAIRepository) GetLeadAIByNiche(userID, niche string) ([]models.LeadAI, error) {
	// Trim whitespace from niche to avoid matching issues
	niche = strings.TrimSpace(niche)
	
	logrus.Debugf("GetLeadAIByNiche - UserID: %s, Niche: '%s' (len=%d)", userID, niche, len(niche))
	
	query := `

		SELECT id, user_id, device_id, name, phone, email, niche, source, 
		       status, target_status, notes, assigned_at, sent_at, created_at, updated_at
		FROM leads_ai
		WHERE user_id = ? AND niche LIKE CONCAT('%', ?, '%')
		ORDER BY created_at DESC`
	
	return r.getLeadAIList(query, userID, niche)
}
func (r *leadAIRepository) GetLeadAIByNicheAndStatus(userID, niche, targetStatus string) ([]models.LeadAI, error) {
	// Trim whitespace from niche to avoid matching issues
	niche = strings.TrimSpace(niche)
	
	logrus.Debugf("GetLeadAIByNicheAndStatus - UserID: %s, Niche: '%s' (len=%d), TargetStatus: %s", 
		userID, niche, len(niche), targetStatus)
	
	var query string
	var args []interface{}
	
	if targetStatus == "all" {
		query = `
			SELECT id, user_id, device_id, name, phone, email, niche, source, 
			       status, target_status, notes, assigned_at, sent_at, created_at, updated_at
			FROM leads_ai
			WHERE user_id = ? AND niche LIKE CONCAT('%', ?, '%') AND status = 'pending'
			order BY created_at ASC`
		args = []interface{}{userID, niche}
	} else {
		query = `
			SELECT id, user_id, device_id, name, phone, email, niche, source, 
			       status, target_status, notes, assigned_at, sent_at, created_at, updated_at
			FROM leads_ai
			WHERE user_id = ? AND niche LIKE CONCAT('%', ?, '%') AND target_status = ? AND status = 'pending'
			order BY created_at ASC`
		args = []interface{}{userID, niche, targetStatus}
	}
	
	return r.getLeadAIListWithArgs(query, args...)
}

func (r *leadAIRepository) AssignDevice(leadID int, deviceID string) error {
	now := time.Now()
	query := `

		UPDATE leads_ai 
		SET device_id = ?, assigned_at = ?, status = 'assigned', updated_at = ?
		WHERE id = ?`
	
	_, err := r.db.Exec(query, deviceID, now, now, leadID)
	return err
}
func (r *leadAIRepository) UpdateStatus(leadID int, status string) error {
	now := time.Now()
	query := `

		UPDATE leads_ai SET status = ?, updated_at = ?`
	
	args := []interface{}{status, now}
	
	if status == "sent" {
		query += `, sent_at = ? WHERE id = ?`
		args = append(args, now, leadID)
	} else {
		query += ` WHERE id = ?`
		args = append(args, leadID)
	}
	
	_, err := r.db.Exec(query, args...)
	return err
}

func (r *leadAIRepository) UpdateLeadAI(id int, lead *models.LeadAI) error {
	query := `

		UPDATE leads_ai 
		SET name = ?, phone = ?, email = ?, niche = ?, 
		    target_status = ?, notes = ?, updated_at = ?
		WHERE id = ?`
	
	_, err := r.db.Exec(
		query,
		lead.Name,
		lead.Phone,
		lead.Email,
		lead.Niche,
		lead.TargetStatus,
		lead.Notes,
		time.Now(),
		id,
	)
	
	return err
}
func (r *leadAIRepository) DeleteLeadAI(id int) error {
	query := `DELETE FROM leads_ai WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *leadAIRepository) GetLeadAICountByDevice(campaignID int, deviceID string) (int, error) {
	var count int
	query := `

		SELECT COUNT(*) 
		FROM ai_campaign_progress 
		WHERE campaign_id = ? AND device_id = ?`
	
	err := r.db.QueryRow(query, campaignID, deviceID).Scan(&count)
	if err != nil {
		return 0, err
	}
	
	return count, nil
}

func (r *leadAIRepository) GetCampaignProgress(campaignID int) ([]models.AICampaignProgress, error) {
	query := `

		SELECT id, campaign_id, device_id, leads_sent, leads_failed, 
		       status, last_activity, created_at, updated_at
		FROM ai_campaign_progress
		WHERE campaign_id = ?
		ORDER BY device_id`
	
	rows, err := r.db.Query(query, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var progresses []models.AICampaignProgress
	for rows.Next() {
		var p models.AICampaignProgress
		err := rows.Scan(
			&p.ID,
			&p.CampaignID,
			&p.DeviceID,
			&p.LeadsSent,
			&p.LeadsFailed,
			&p.Status,
			&p.LastActivity,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			continue
		}
		progresses = append(progresses, p)
	}
	
	return progresses, nil
}
func (r *leadAIRepository) UpdateCampaignProgress(progress *models.AICampaignProgress) error {
	query := `

		INSERT INTO ai_campaign_progress(campaign_id, device_id, leads_sent, leads_failed, status)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (campaign_id, device_id) 
		DO UPDATE SET 
			leads_sent = ?,
			leads_failed = ?,
			status = ?,
			last_activity = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP`
	
	_, err := r.db.Exec(
		query,
		progress.CampaignID,
		progress.DeviceID,
		progress.LeadsSent,
		progress.LeadsFailed,
		progress.Status,
	)
	
	return err
}

// Helper methods
func (r *leadAIRepository) getLeadAIList(query string, args ...interface{}) ([]models.LeadAI, error) {
	return r.getLeadAIListWithArgs(query, args...)
}

func (r *leadAIRepository) getLeadAIListWithArgs(query string, args ...interface{}) ([]models.LeadAI, error) {
	logrus.Debugf("getLeadAIListWithArgs - Query: %s, Args: %v", query, args)
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var leads []models.LeadAI
	var count int
	for rows.Next() {
		var lead models.LeadAI
		err := rows.Scan(
			&lead.ID,
			&lead.UserID,
			&lead.DeviceID,
			&lead.Name,
			&lead.Phone,
			&lead.Email,
			&lead.Niche,
			&lead.Source,
			&lead.Status,
			&lead.TargetStatus,
			&lead.Notes,
			&lead.AssignedAt,
			&lead.SentAt,
			&lead.CreatedAt,
			&lead.UpdatedAt,
		)
		if err != nil {
			continue
		}
		
		// Debug: Show what niche was found
		if len(args) >= 2 {
			searchNiche, ok := args[1].(string)
			if ok {
				logrus.Debugf("Found AI lead with niche: '%s' (searching for '%s')", lead.Niche, searchNiche)
			}
		}
		
		leads = append(leads, lead)
		count++
	}
	
	logrus.Debugf("getLeadAIListWithArgs - Found %d AI leads", count)
	
	return leads, nil
}
