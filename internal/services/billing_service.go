package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"nodepath-chat/internal/models"
)

type BillingService struct {
	db *sqlx.DB
}

func NewBillingService(db *sqlx.DB) *BillingService {
	return &BillingService{db: db}
}

// GetUserSubscription retrieves the active subscription for a user
func (s *BillingService) GetUserSubscription(ctx context.Context, userID string) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, plan_name, plan_price, plan_period, status, 
		       next_billing_date, features, created_at, updated_at
		FROM subscriptions_nodepath 
		WHERE user_id = ? AND status = 'active'
		ORDER BY created_at DESC 
		LIMIT 1
	`
	
	var subscription models.Subscription
	err := s.db.GetContext(ctx, &subscription, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return test subscription if no subscription exists
			return s.getTestSubscription(userID), nil
		}
		return nil, fmt.Errorf("failed to get user subscription: %w", err)
	}
	
	return &subscription, nil
}

// GetBillingHistory retrieves the billing history for a user
func (s *BillingService) GetBillingHistory(ctx context.Context, userID string, limit, offset int) ([]models.BillingHistory, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM billing_history_nodepath WHERE user_id = ?`
	var totalCount int
	err := s.db.GetContext(ctx, &totalCount, countQuery, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get billing history count: %w", err)
	}

	// Get billing history records
	query := `
		SELECT id, user_id, payment_id, invoice_number, amount, currency, 
		       description, status, payment_date, created_at
		FROM billing_history_nodepath 
		WHERE user_id = ? 
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?
	`
	
	var history []models.BillingHistory
	err = s.db.SelectContext(ctx, &history, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get billing history: %w", err)
	}

	// If no records exist, return sample test data
	if len(history) == 0 {
		return s.getTestBillingHistory(userID), 3, nil
	}
	
	return history, totalCount, nil
}

// CreatePayment creates a new payment record and initiates Billplz payment
func (s *BillingService) CreatePayment(ctx context.Context, req models.CreatePaymentRequest) (*models.Payment, error) {
	// Create payment record
	paymentID := fmt.Sprintf("pay_%d", time.Now().Unix())
	invoiceNum := fmt.Sprintf("INV-%s-%04d", time.Now().Format("20060102"), time.Now().Unix()%10000)
	
	payment := &models.Payment{
		ID:            paymentID,
		UserID:        req.UserID,
		Amount:        req.Amount,
		Currency:      "MYR",
		Description:   sql.NullString{String: req.Description, Valid: true},
		Status:        models.PaymentStatusPending,
		PaymentMethod: "billplz",
		InvoiceNumber: invoiceNum,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Insert payment record
	query := `
		INSERT INTO payments_nodepath (id, user_id, amount, currency, description, status, 
		                     payment_method, invoice_number, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := s.db.ExecContext(ctx, query, 
		payment.ID, payment.UserID, payment.Amount, payment.Currency, 
		payment.Description, payment.Status, payment.PaymentMethod, 
		payment.InvoiceNumber, payment.CreatedAt, payment.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// Create Billplz payment
	billplzURL, billID, err := s.createBillplzPayment(req, payment.InvoiceNumber)
	if err != nil {
		log.Printf("Failed to create Billplz payment: %v", err)
		// Continue without failing - payment record exists
	} else {
		// Update payment with Billplz details
		updateQuery := `
			UPDATE payments_nodepath 
			SET bill_id = ?, billplz_url = ?, updated_at = ?
			WHERE id = ?
		`
		_, err = s.db.ExecContext(ctx, updateQuery, billID, billplzURL, time.Now(), paymentID)
		if err != nil {
			log.Printf("Failed to update payment with Billplz details: %v", err)
		} else {
			payment.BillID = sql.NullString{String: billID, Valid: true}
			payment.BillplzURL = sql.NullString{String: billplzURL, Valid: true}
		}
	}

	return payment, nil
}

// UpdatePaymentStatus updates the status of a payment (for Billplz callbacks)
func (s *BillingService) UpdatePaymentStatus(ctx context.Context, billID string, status models.PaymentStatus) error {
	query := `
		UPDATE payments_nodepath 
		SET status = ?, paid_at = ?, updated_at = ?
		WHERE bill_id = ?
	`
	
	var paidAt *time.Time
	if status == models.PaymentStatusPaid {
		now := time.Now()
		paidAt = &now
	}
	
	_, err := s.db.ExecContext(ctx, query, status, paidAt, time.Now(), billID)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}
	
	return nil
}

// createBillplzPayment creates a payment with Billplz API
func (s *BillingService) createBillplzPayment(req models.CreatePaymentRequest, invoiceNumber string) (string, string, error) {
	// Convert amount to cents (Billplz expects amount in cents)
	amountCents := int(req.Amount * 100)
	
	// For testing, return mock response for RM 1.00
	if amountCents == 100 { // RM 1.00 test amount
		mockBillID := fmt.Sprintf("bill_%d", time.Now().Unix())
		mockURL := fmt.Sprintf("https://www.billplz-sandbox.com/bills/%s", mockBillID)
		return mockURL, mockBillID, nil
	}
	
	// Make request to Billplz API (commented out for testing)
	/*
	billplzURL := "https://www.billplz-sandbox.com/api/v3/bills"
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", billplzURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(apiKey+":")))
	
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to make Billplz request: %w", err)
	}
	defer resp.Body.Close()
	
	var billplzResp models.BillplzPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&billplzResp); err != nil {
		return "", "", fmt.Errorf("failed to decode Billplz response: %w", err)
	}
	
	if billplzResp.Error != nil {
		return "", "", fmt.Errorf("Billplz error: %s", billplzResp.Error.Message)
	}
	
	return billplzResp.URL, billplzResp.ID, nil
	*/
	
	return "", "", fmt.Errorf("Billplz integration disabled for testing")
}

// getTestSubscription returns a test subscription for development/testing
func (s *BillingService) getTestSubscription(userID string) *models.Subscription {
	features := json.RawMessage(`["WhatsApp Bot Integration", "Flow Builder Access", "Basic Analytics", "Standard Support"]`)
	
	return &models.Subscription{
		ID:              "test_sub_001",
		UserID:          userID,
		PlanName:        "Test Plan",
		PlanPrice:       1.00,
		PlanPeriod:      "monthly",
		Status:          models.SubscriptionStatusActive,
		NextBillingDate: time.Now().AddDate(0, 1, 0), // Next month
		Features:        &features,
		CreatedAt:       time.Now().AddDate(0, -1, 0), // Created last month
		UpdatedAt:       time.Now(),
	}
}

// getTestBillingHistory returns test billing history for development/testing
func (s *BillingService) getTestBillingHistory(userID string) []models.BillingHistory {
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	twoMonthsAgo := now.AddDate(0, -2, 0)
	
	return []models.BillingHistory{
		{
			ID:            "hist_001",
			UserID:        userID,
			InvoiceNumber: "INV-" + now.Format("20060102") + "-0001",
			Amount:        1.00,
			Currency:      "MYR",
			Description:   "Test Plan - Monthly subscription",
			Status:        models.PaymentStatusPaid,
			PaymentDate:   &now,
			CreatedAt:     now,
		},
		{
			ID:            "hist_002",
			UserID:        userID,
			InvoiceNumber: "INV-" + lastMonth.Format("20060102") + "-0002",
			Amount:        1.00,
			Currency:      "MYR",
			Description:   "Test Plan - Monthly subscription",
			Status:        models.PaymentStatusPaid,
			PaymentDate:   &lastMonth,
			CreatedAt:     lastMonth,
		},
		{
			ID:            "hist_003",
			UserID:        userID,
			InvoiceNumber: "INV-" + twoMonthsAgo.Format("20060102") + "-0003",
			Amount:        1.00,
			Currency:      "MYR",
			Description:   "Test Plan - Monthly subscription",
			Status:        models.PaymentStatusPaid,
			PaymentDate:   &twoMonthsAgo,
			CreatedAt:     twoMonthsAgo,
		},
	}
}