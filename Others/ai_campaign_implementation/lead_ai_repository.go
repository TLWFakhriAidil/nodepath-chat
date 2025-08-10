package repository

import (
	"database/sql"
	"fmt"
	"time"
	
	"go-whatsapp-web-multidevice/config"
	"go-whatsapp-web-multidevice/models"
	"go-whatsapp-web-multidevice/utils"
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
	log utils.Logger
}

var leadAIRepo *leadAIRepository

func GetLeadAIRepository() LeadAIRepository {
	if leadAIRepo == nil {
		leadAIRepo = &leadAIRepository{
			db:  config.GetDB(),
			log: utils.GetLogger(),
		}
	}
	return leadAIRepo
}

func (r *leadAIRepository) CreateLeadAI(lead *models.LeadAI) error {
	query := `
		INSERT INTO leads_ai (user_id, name, phone, email, niche, source, target_status, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`
	
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
	
	if err != nil {
		r.log.Errorf("Failed to create AI lead: %v", err)
		return err
	}
	
	return nil
}

func (r *leadAIRepository) GetLeadAIByID(id int) (*models.LeadAI, error) {
	var lead models.LeadAI
	query := `
		SELECT id, user_id, device_id, name, phone, email, niche, source, 
		       status, target_status, notes, assigned_at, sent_at, created_at, updated_at
		FROM leads_ai
		WHERE id = $1`
	
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
		WHERE user_id = $1
		ORDER BY created_at DESC`
	
	return r.getLeadAIList(query, userID)
}
func (r *leadAIRepository) GetPendingLeadAI(userID string) ([]models.LeadAI, error) {
	query := `
		SELECT id, user_id, device_id, name, phone, email, niche, source, 
		       status, target_status, notes, assigned_at, sent_at, created_at, updated_at
		FROM leads_ai
		WHERE user_id = $1 AND status = 'pending'
		ORDER BY created_at ASC`
	
	return r.getLeadAIList(query, userID)
}

func (r *leadAIRepository) GetLeadAIByNiche(userID, niche string) ([]models.LeadAI, error) {
	query := `
		SELECT id, user_id, device_id, name, phone, email, niche, source, 
		       status, target_status, notes, assigned_at, sent_at, created_at, updated_at
		FROM leads_ai
		WHERE user_id = $1 AND niche = $2
		ORDER BY created_at DESC`
	
	return r.getLeadAIList(query, userID, niche)
}

func (r *leadAIRepository) GetLeadAIByNicheAndStatus(userID, niche, targetStatus string) ([]models.LeadAI, error) {
	var query string
	var args []interface{}
	
	if targetStatus == "all" {
		query = `
			SELECT id, user_id, device_id, name, phone, email, niche, source, 
			       status, target_status, notes, assigned_at, sent_at, created_at, updated_at
			FROM leads_ai
			WHERE user_id = $1 AND niche = $2 AND status = 'pending'
			ORDER BY created_at ASC`
		args = []interface{}{userID, niche}
	} else {
		query = `
			SELECT id, user_id, device_id, name, phone, email, niche, source, 
			       status, target_status, notes, assigned_at, sent_at, created_at, updated_at
			FROM leads_ai
			WHERE user_id = $1 AND niche = $2 AND target_status = $3 AND status = 'pending'
			ORDER BY created_at ASC`
		args = []interface{}{userID, niche, targetStatus}
	}
	
	return r.getLeadAIListWithArgs(query, args...)
}
func (r *leadAIRepository) AssignDevice(leadID int, deviceID string) error {
	now := time.Now()
	query := `
		UPDATE leads_ai 
		SET device_id = $1, assigned_at = $2, status = 'assigned', updated_at = $3
		WHERE id = $4`
	
	_, err := r.db.Exec(query, deviceID, now, now, leadID)
	if err != nil {
		r.log.Errorf("Failed to assign device to AI lead: %v", err)
		return err
	}
	
	return nil
}

func (r *leadAIRepository) UpdateStatus(leadID int, status string) error {
	now := time.Now()
	query := `
		UPDATE leads_ai 
		SET status = $1, updated_at = $2`
	
	args := []interface{}{status, now}
	
	// If status is "sent", also update sent_at
	if status == "sent" {
		query += `, sent_at = $3 WHERE id = $4`
		args = append(args, now, leadID)
	} else {
		query += ` WHERE id = $3`
		args = append(args, leadID)
	}
	
	_, err := r.db.Exec(query, args...)
	if err != nil {
		r.log.Errorf("Failed to update AI lead status: %v", err)
		return err
	}
	
	return nil
}
func (r *leadAIRepository) UpdateLeadAI(id int, lead *models.LeadAI) error {
	query := `
		UPDATE leads_ai 
		SET name = $1, phone = $2, email = $3, niche = $4, 
		    target_status = $5, notes = $6, updated_at = $7
		WHERE id = $8`
	
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
	
	if err != nil {
		r.log.Errorf("Failed to update AI lead: %v", err)
		return err
	}
	
	return nil
}

func (r *leadAIRepository) DeleteLeadAI(id int) error {
	query := `DELETE FROM leads_ai WHERE id = $1`
	
	_, err := r.db.Exec(query, id)
	if err != nil {
		r.log.Errorf("Failed to delete AI lead: %v", err)
		return err
	}
	
	return nil
}
func (r *leadAIRepository) GetLeadAICountByDevice(campaignID int, deviceID string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) 
		FROM ai_campaign_progress 
		WHERE campaign_id = $1 AND device_id = $2`
	
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
		WHERE campaign_id = $1
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
			r.log.Errorf("Failed to scan campaign progress: %v", err)
			continue
		}
		progresses = append(progresses, p)
	}
	
	return progresses, nil
}
func (r *leadAIRepository) UpdateCampaignProgress(progress *models.AICampaignProgress) error {
	query := `
		INSERT INTO ai_campaign_progress (campaign_id, device_id, leads_sent, leads_failed, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (campaign_id, device_id) 
		DO UPDATE SET 
			leads_sent = $3,
			leads_failed = $4,
			status = $5,
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
	
	if err != nil {
		r.log.Errorf("Failed to update campaign progress: %v", err)
		return err
	}
	
	return nil
}

// Helper methods
func (r *leadAIRepository) getLeadAIList(query string, args ...interface{}) ([]models.LeadAI, error) {
	return r.getLeadAIListWithArgs(query, args...)
}

func (r *leadAIRepository) getLeadAIListWithArgs(query string, args ...interface{}) ([]models.LeadAI, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var leads []models.LeadAI
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
			r.log.Errorf("Failed to scan AI lead: %v", err)
			continue
		}
		leads = append(leads, lead)
	}
	
	return leads, nil
}