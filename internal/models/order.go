package models

import (
	"time"
)

// Order represents a billing order in the system
type Order struct {
	ID           int       `json:"id" db:"id"`
	Amount       float64   `json:"amount" db:"amount"` // Amount in RM
	CollectionID *string   `json:"collection_id" db:"collection_id"`
	Status       string    `json:"status" db:"status"` // Pending, Processing, Success, Failed
	BillID       *string   `json:"bill_id" db:"bill_id"`
	URL          *string   `json:"url" db:"url"` // Billplz payment URL
	Product      string    `json:"product" db:"product"`
	Method       string    `json:"method" db:"method"`   // billplz only
	UserID       *string   `json:"user_id" db:"user_id"` // CHAR(36) UUID from user_nodepath
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	Amount  float64 `json:"amount" binding:"required"` // Amount in RM
	Product string  `json:"product" binding:"required"`
}

// BillplzCreateBillRequest represents request to Billplz API
type BillplzCreateBillRequest struct {
	CollectionID    string `json:"collection_id"`
	Email           string `json:"email"`
	Name            string `json:"name"`
	Amount          int    `json:"amount"` // Amount in cents (sen)
	Description     string `json:"description"`
	CallbackURL     string `json:"callback_url"`
	RedirectURL     string `json:"redirect_url"`
	Reference1      string `json:"reference_1"`
	Reference1Label string `json:"reference_1_label"`
}

// BillplzCreateBillResponse represents response from Billplz API
type BillplzCreateBillResponse struct {
	ID              string `json:"id"`
	CollectionID    string `json:"collection_id"`
	Paid            bool   `json:"paid"`
	State           string `json:"state"`
	Amount          int    `json:"amount"`
	PaidAmount      int    `json:"paid_amount"`
	DueAt           string `json:"due_at"`
	Email           string `json:"email"`
	Mobile          string `json:"mobile"`
	Name            string `json:"name"`
	URL             string `json:"url"`
	Reference1Label string `json:"reference_1_label"`
	Reference1      string `json:"reference_1"`
	Description     string `json:"description"`
}

// BillplzCallbackData represents callback data from Billplz
type BillplzCallbackData struct {
	ID              string `form:"id"`
	CollectionID    string `form:"collection_id"`
	Paid            string `form:"paid"` // "true" or "false"
	State           string `form:"state"`
	Amount          string `form:"amount"`
	PaidAmount      string `form:"paid_amount"`
	DueAt           string `form:"due_at"`
	Email           string `form:"email"`
	Mobile          string `form:"mobile"`
	Name            string `form:"name"`
	URL             string `form:"url"`
	PaidAt          string `form:"paid_at"`
	Reference1Label string `form:"reference_1_label"`
	Reference1      string `form:"reference_1"`
	Description     string `form:"description"`
}
