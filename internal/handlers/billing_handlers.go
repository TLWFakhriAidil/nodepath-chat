package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/services"
)

type BillingHandlers struct {
	billingService *services.BillingService
}

func NewBillingHandlers(billingService *services.BillingService) *BillingHandlers {
	return &BillingHandlers{
		billingService: billingService,
	}
}

// GetBillingData gets subscription and billing history for the authenticated user
func (h *BillingHandlers) GetBillingData(c *fiber.Ctx) error {
	// For testing, use a test user ID - in production, get from JWT token
	userID := "00000000-0000-0000-0000-000000000001"
	
	// Parse query parameters
	limitStr := c.Query("limit", "10")
	offsetStr := c.Query("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	
	// Get subscription
	subscription, err := h.billingService.GetUserSubscription(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get subscription data",
			"details": err.Error(),
		})
	}
	
	// Get billing history
	history, totalCount, err := h.billingService.GetBillingHistory(c.Context(), userID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get billing history",
			"details": err.Error(),
		})
	}
	
	response := models.BillingResponse{
		Subscription:   subscription,
		BillingHistory: history,
		TotalCount:     totalCount,
	}
	
	return c.JSON(response)
}

// GetSubscription gets the current subscription for the authenticated user
func (h *BillingHandlers) GetSubscription(c *fiber.Ctx) error {
	// For testing, use a test user ID - in production, get from JWT token
	userID := "00000000-0000-0000-0000-000000000001"
	
	subscription, err := h.billingService.GetUserSubscription(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get subscription",
			"details": err.Error(),
		})
	}
	
	return c.JSON(subscription)
}

// GetBillingHistoryOnly gets the billing history for the authenticated user
func (h *BillingHandlers) GetBillingHistoryOnly(c *fiber.Ctx) error {
	// For testing, use a test user ID - in production, get from JWT token
	userID := "00000000-0000-0000-0000-000000000001"
	
	// Parse query parameters
	limitStr := c.Query("limit", "10")
	offsetStr := c.Query("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	
	history, totalCount, err := h.billingService.GetBillingHistory(c.Context(), userID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get billing history",
			"details": err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"data": history,
		"total_count": totalCount,
		"limit": limit,
		"offset": offset,
	})
}

// CreatePayment creates a new payment and initiates Billplz payment
func (h *BillingHandlers) CreatePayment(c *fiber.Ctx) error {
	// For testing, use a test user ID - in production, get from JWT token
	userID := "00000000-0000-0000-0000-000000000001"
	
	var req models.CreatePaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}
	
	// Set user ID from context (for testing, hardcoded)
	req.UserID = userID
	
	// Set default values if not provided
	if req.Amount <= 0 {
		req.Amount = 1.00 // Default test amount
	}
	if req.Description == "" {
		req.Description = "Test Plan - Monthly subscription"
	}
	if req.CustomerEmail == "" {
		req.CustomerEmail = "test@example.com"
	}
	if req.CustomerName == "" {
		req.CustomerName = "Test Customer"
	}
	
	payment, err := h.billingService.CreatePayment(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create payment",
			"details": err.Error(),
		})
	}
	
	return c.Status(fiber.StatusCreated).JSON(payment)
}

// BillplzCallback handles Billplz payment callbacks
func (h *BillingHandlers) BillplzCallback(c *fiber.Ctx) error {
	// Parse Billplz callback data
	var callbackData map[string]interface{}
	if err := c.BodyParser(&callbackData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid callback data",
			"details": err.Error(),
		})
	}
	
	// Extract bill ID and payment status from callback
	billID, ok := callbackData["id"].(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing bill ID in callback",
		})
	}
	
	paid, ok := callbackData["paid"].(bool)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing payment status in callback",
		})
	}
	
	// Determine payment status
	var status models.PaymentStatus
	if paid {
		status = models.PaymentStatusPaid
	} else {
		status = models.PaymentStatusFailed
	}
	
	// Update payment status
	err := h.billingService.UpdatePaymentStatus(c.Context(), billID, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update payment status",
			"details": err.Error(),
		})
	}
	
	return c.JSON(fiber.Map{
		"message": "Payment status updated successfully",
		"bill_id": billID,
		"status": status,
	})
}

// TestPayment creates a test payment for RM 1.00 to test Billplz integration
func (h *BillingHandlers) TestPayment(c *fiber.Ctx) error {
	// For testing, use a test user ID - in production, get from JWT token
	userID := "00000000-0000-0000-0000-000000000001"
	
	req := models.CreatePaymentRequest{
		UserID:        userID,
		Amount:        1.00,
		Description:   "Test Plan - Monthly subscription (RM 1.00 test)",
		CustomerEmail: "test@example.com",
		CustomerName:  "Test Customer",
	}
	
	payment, err := h.billingService.CreatePayment(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create test payment",
			"details": err.Error(),
		})
	}
	
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Test payment created successfully",
		"payment": payment,
		"test_note": "This is a RM 1.00 test payment for Billplz integration testing",
	})
}