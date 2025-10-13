package handlers

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"
	"nodepath-chat/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// BillingHandlers handles billing and payment operations
type BillingHandlers struct {
	orderRepo      repository.OrderRepository
	billplzService *services.BillplzService
	db             *sql.DB
}

// NewBillingHandlers creates a new billing handlers instance
func NewBillingHandlers(orderRepo repository.OrderRepository, billplzService *services.BillplzService, db *sql.DB) *BillingHandlers {
	return &BillingHandlers{
		orderRepo:      orderRepo,
		billplzService: billplzService,
		db:             db,
	}
}

// CreateOrder handles order creation and payment processing
func (h *BillingHandlers) CreateOrder(c *fiber.Ctx) error {
	var req models.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get user ID from context (required) - CHAR(36) UUID string
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Check user profile for gmail, phone, and full_name from user_nodepath table
	var gmail, phone, fullName sql.NullString
	query := `SELECT gmail, phone, full_name FROM user_nodepath WHERE id = ?`
	err := h.db.QueryRow(query, userID).Scan(&gmail, &phone, &fullName)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "profile_incomplete",
				"message": "Profile not found. Please complete your profile first.",
			})
		}
		logrus.WithError(err).Error("Failed to check user profile")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check profile",
		})
	}

	// Check if gmail, phone, or full_name is null or empty
	if !gmail.Valid || gmail.String == "" || !phone.Valid || phone.String == "" || !fullName.Valid || fullName.String == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "profile_incomplete",
			"message": "Please update your profile with email, phone number, and full name before upgrading.",
		})
	}

	// Create order in database with collection_id set
	collectionID := h.billplzService.GetCollectionID()
	order := &models.Order{
		Amount:       req.Amount,
		CollectionID: &collectionID,
		Status:       "Processing",
		BillID:       nil,
		URL:          nil,
		Product:      req.Product,
		Method:       "billplz",
		UserID:       &userID,
	}

	orderID, err := h.orderRepo.CreateOrder(order)
	if err != nil {
		logrus.WithError(err).Error("Failed to create order")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create order",
		})
	}

	// For Billplz payment, create bill
	// Priority: BASE_URL (set in Railway) > RAILWAY_PUBLIC_DOMAIN > localhost
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		// Try Railway public domain
		domain := os.Getenv("RAILWAY_PUBLIC_DOMAIN")
		if domain != "" {
			baseURL = "https://" + domain
		} else {
			baseURL = "http://localhost:8080" // Default for local development
		}
	}

	billReq := models.BillplzCreateBillRequest{
		CollectionID:    h.billplzService.GetCollectionID(),
		Email:           gmail.String,
		Name:            fullName.String,
		Amount:          h.billplzService.ConvertRMToSen(req.Amount), // Convert RM to sen
		Description:     req.Product,
		CallbackURL:     fmt.Sprintf("%s/api/billing/callback", baseURL),
		RedirectURL:     fmt.Sprintf("%s/billings?order_id=%d", baseURL, orderID),
		Reference1:      req.Product, // Show product name instead of order ID
		Reference1Label: "Package",   // Label as "Package" instead of "Order ID"
	}

	billResp, err := h.billplzService.CreateBill(billReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to create Billplz bill")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create payment bill",
		})
	}

	// Update order with bill information
	err = h.orderRepo.UpdateOrderBillInfo(orderID, billResp.ID, billResp.URL)
	if err != nil {
		logrus.WithError(err).Error("Failed to update order with bill info")
		// Don't fail the request, bill is already created
	}

	return c.JSON(fiber.Map{
		"order_id":       orderID,
		"bill_id":        billResp.ID,
		"payment_url":    billResp.URL,
		"payment_method": "billplz",
		"status":         "Processing",
		"message":        "Please complete payment via the provided URL",
	})
}

// BillplzCallback handles payment callback from Billplz
// Supports both POST (webhook) and GET (redirect) callbacks
func (h *BillingHandlers) BillplzCallback(c *fiber.Ctx) error {
	var callback models.BillplzCallbackData

	// Try to parse as POST body first (webhook callback)
	if c.Method() == "POST" {
		if err := c.BodyParser(&callback); err != nil {
			logrus.WithError(err).Error("Failed to parse POST callback data")
			return c.SendStatus(fiber.StatusBadRequest)
		}
	} else {
		// Parse as GET query params (redirect callback)
		callback.ID = c.Query("billplz[id]")
		callback.CollectionID = c.Query("billplz[collection_id]")
		callback.Paid = c.Query("billplz[paid]")
		callback.State = c.Query("billplz[state]")
		callback.Amount = c.Query("billplz[amount]")
		callback.PaidAmount = c.Query("billplz[paid_amount]")
		callback.DueAt = c.Query("billplz[due_at]")
		callback.Email = c.Query("billplz[email]")
		callback.Mobile = c.Query("billplz[mobile]")
		callback.Name = c.Query("billplz[name]")
		callback.URL = c.Query("billplz[url]")
		callback.PaidAt = c.Query("billplz[paid_at]")
		callback.Reference1Label = c.Query("billplz[reference_1_label]")
		callback.Reference1 = c.Query("billplz[reference_1]")
		callback.Description = c.Query("billplz[description]")
	}

	logrus.WithFields(logrus.Fields{
		"bill_id": callback.ID,
		"paid":    callback.Paid,
		"amount":  callback.Amount,
		"method":  c.Method(),
	}).Info("Received Billplz callback")

	// Check if payment was successful
	if callback.Paid == "true" {
		// Update order status to Success
		err := h.orderRepo.UpdateOrderStatus(callback.ID, "Success")
		if err != nil {
			logrus.WithError(err).Error("Failed to update order status")
			if c.Method() == "POST" {
				return c.SendStatus(fiber.StatusInternalServerError)
			}
			// For GET redirect, still redirect to billings page
		} else {
			logrus.WithField("bill_id", callback.ID).Info("Payment successful, order updated")

			// Update user subscription in user_nodepath
			// Get order to find user_id
			order, err := h.orderRepo.GetOrderByBillID(callback.ID)
			if err == nil && order != nil && order.UserID != nil {
				// Update user status to "Pro" and set expired date to now + 1 month
				updateQuery := `
					UPDATE user_nodepath 
					SET status = 'Pro', 
					    expired = DATE_ADD(NOW(), INTERVAL 1 MONTH)
					WHERE id = ?
				`
				_, err = h.db.Exec(updateQuery, *order.UserID)
				if err != nil {
					logrus.WithError(err).WithField("user_id", *order.UserID).Error("Failed to update user subscription")
				} else {
					logrus.WithField("user_id", *order.UserID).Info("User subscription updated to Pro for 1 month")
				}
			}
		}
	} else {
		// Payment not successful
		err := h.orderRepo.UpdateOrderStatus(callback.ID, "Failed")
		if err != nil {
			logrus.WithError(err).Warn("Failed to update order status to Failed")
		}
		logrus.WithField("bill_id", callback.ID).Warn("Payment not successful")
	}

	// For POST webhook, respond with OK
	if c.Method() == "POST" {
		return c.SendString("OK")
	}

	// For GET redirect, redirect to billings page
	orderID := c.Query("order_id")
	if orderID != "" {
		return c.Redirect("/billings?order_id=" + orderID)
	}
	return c.Redirect("/billings")
}

// GetOrder retrieves order details
func (h *BillingHandlers) GetOrder(c *fiber.Ctx) error {
	orderIDStr := c.Params("id")
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid order ID",
		})
	}

	order, err := h.orderRepo.GetOrderByID(orderID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get order")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get order",
		})
	}

	if order == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	return c.JSON(order)
}

// GetOrders retrieves orders for current user or all orders (admin)
func (h *BillingHandlers) GetOrders(c *fiber.Ctx) error {
	// Get pagination parameters
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	// Get user ID from context - CHAR(36) UUID string
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Get orders for user
	orders, total, err := h.orderRepo.GetOrdersByUserID(userID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to get orders")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get orders",
		})
	}

	return c.JSON(fiber.Map{
		"orders": orders,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetAllOrders retrieves all orders (admin only)
func (h *BillingHandlers) GetAllOrders(c *fiber.Ctx) error {
	// Get pagination parameters
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	// Get all orders
	orders, total, err := h.orderRepo.GetAllOrders(limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to get all orders")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get orders",
		})
	}

	return c.JSON(fiber.Map{
		"orders": orders,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
