package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"nodepath-chat/internal/models"

	"github.com/sirupsen/logrus"
)

// BillplzService handles Billplz payment API integration
type BillplzService struct {
	apiKey       string
	collectionID string
	baseURL      string
}

// NewBillplzService creates a new Billplz service instance
// Using the API key and collection ID from the PHP code provided
func NewBillplzService() *BillplzService {
	return &BillplzService{
		apiKey:       "948edf23-8c36-45fc-8457-1eb77b649616", // From PHP code
		collectionID: "watojri1",                             // From PHP code
		baseURL:      "https://www.billplz.com/api/v3",
	}
}

// CreateBill creates a new bill in Billplz
func (s *BillplzService) CreateBill(req models.BillplzCreateBillRequest) (*models.BillplzCreateBillResponse, error) {
	url := fmt.Sprintf("%s/bills", s.baseURL)

	// Convert struct to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.apiKey+":")))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check if request was successful
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		logrus.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"response":    string(body),
		}).Error("Billplz API error")
		return nil, fmt.Errorf("billplz API error: %s", string(body))
	}

	// Parse response
	var billResp models.BillplzCreateBillResponse
	if err := json.Unmarshal(body, &billResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"bill_id": billResp.ID,
		"amount":  billResp.Amount,
		"email":   billResp.Email,
	}).Info("Billplz bill created successfully")

	return &billResp, nil
}

// GetCollectionID returns the collection ID
func (s *BillplzService) GetCollectionID() string {
	return s.collectionID
}

// ConvertRMToSen converts RM to sen (cents)
func (s *BillplzService) ConvertRMToSen(rm float64) int {
	return int(rm * 100)
}

// ConvertSenToRM converts sen (cents) to RM
func (s *BillplzService) ConvertSenToRM(sen int) float64 {
	return float64(sen) / 100
}

// ParseAmount parses amount string to integer (sen)
func (s *BillplzService) ParseAmount(amountStr string) (int, error) {
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse amount: %w", err)
	}
	return amount, nil
}
