package repository

import (
	"database/sql"
	"fmt"
	"time"

	"nodepath-chat/internal/models"

	"github.com/sirupsen/logrus"
)

// OrderRepository interface defines methods for order management
type OrderRepository interface {
	CreateOrder(order *models.Order) (int, error)
	GetOrderByID(id int) (*models.Order, error)
	GetOrderByBillID(billID string) (*models.Order, error)
	UpdateOrderStatus(billID string, status string) error
	UpdateOrderBillInfo(orderID int, billID string, url string) error
	GetOrdersByUserID(userID string, limit int, offset int) ([]models.Order, int, error)
	GetAllOrders(limit int, offset int) ([]models.Order, int, error)
}

// orderRepository implements OrderRepository interface
type orderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new instance of OrderRepository
func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{
		db: db,
	}
}

// CreateOrder creates a new order in the database
func (r *orderRepository) CreateOrder(order *models.Order) (int, error) {
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	query := `
		INSERT INTO orders_nodepath (
			amount, collection_id, status, bill_id, url, product, method, user_id,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		order.Amount, order.CollectionID, order.Status, order.BillID, order.URL,
		order.Product, order.Method, order.UserID,
		order.CreatedAt, order.UpdatedAt,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get order ID: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"order_id": orderID,
		"amount":   order.Amount,
		"method":   order.Method,
	}).Info("Order created successfully")

	return int(orderID), nil
}

// GetOrderByID retrieves an order by ID
func (r *orderRepository) GetOrderByID(id int) (*models.Order, error) {
	query := `
		SELECT id, amount, collection_id, status, bill_id, url, product, method, user_id,
		       created_at, updated_at
		FROM orders_nodepath
		WHERE id = ?
	`

	var order models.Order
	err := r.db.QueryRow(query, id).Scan(
		&order.ID, &order.Amount, &order.CollectionID, &order.Status, &order.BillID, &order.URL,
		&order.Product, &order.Method, &order.UserID, &order.CreatedAt, &order.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &order, nil
}

// GetOrderByBillID retrieves an order by Billplz bill ID
func (r *orderRepository) GetOrderByBillID(billID string) (*models.Order, error) {
	query := `
		SELECT id, amount, collection_id, status, bill_id, url, product, method, user_id,
		       created_at, updated_at
		FROM orders_nodepath
		WHERE bill_id = ?
	`

	var order models.Order
	err := r.db.QueryRow(query, billID).Scan(
		&order.ID, &order.Amount, &order.CollectionID, &order.Status, &order.BillID, &order.URL,
		&order.Product, &order.Method, &order.UserID, &order.CreatedAt, &order.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order by bill_id: %w", err)
	}

	return &order, nil
}

// UpdateOrderStatus updates the payment status of an order
func (r *orderRepository) UpdateOrderStatus(billID string, status string) error {
	query := `UPDATE orders_nodepath SET status = ?, updated_at = ? WHERE bill_id = ?`

	result, err := r.db.Exec(query, status, time.Now(), billID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found with bill_id: %s", billID)
	}

	logrus.WithFields(logrus.Fields{
		"bill_id": billID,
		"status":  status,
	}).Info("Order status updated successfully")

	return nil
}

// UpdateOrderBillInfo updates the Billplz bill ID and URL for an order
func (r *orderRepository) UpdateOrderBillInfo(orderID int, billID string, url string) error {
	query := `UPDATE orders_nodepath SET bill_id = ?, url = ?, updated_at = ? WHERE id = ?`

	result, err := r.db.Exec(query, billID, url, time.Now(), orderID)
	if err != nil {
		return fmt.Errorf("failed to update order bill info: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found with ID: %d", orderID)
	}

	logrus.WithFields(logrus.Fields{
		"order_id": orderID,
		"bill_id":  billID,
	}).Info("Order bill info updated successfully")

	return nil
}

// GetOrdersByUserID retrieves orders for a specific user
func (r *orderRepository) GetOrdersByUserID(userID string, limit int, offset int) ([]models.Order, int, error) {
	// Get total count
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM orders_nodepath WHERE user_id = ?`
	err := r.db.QueryRow(countQuery, userID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get order count: %w", err)
	}

	// Get orders
	query := `
		SELECT id, amount, collection_id, status, bill_id, url, product, method, user_id,
		       created_at, updated_at
		FROM orders_nodepath
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID, &order.Amount, &order.CollectionID, &order.Status, &order.BillID, &order.URL,
			&order.Product, &order.Method, &order.UserID, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, totalCount, nil
}

// GetAllOrders retrieves all orders (admin use)
func (r *orderRepository) GetAllOrders(limit int, offset int) ([]models.Order, int, error) {
	// Get total count
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM orders_nodepath`
	err := r.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get order count: %w", err)
	}

	// Get orders
	query := `
		SELECT id, amount, collection_id, status, bill_id, url, product, method, user_id,
		       created_at, updated_at
		FROM orders_nodepath
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID, &order.Amount, &order.CollectionID, &order.Status, &order.BillID, &order.URL,
			&order.Product, &order.Method, &order.UserID, &order.CreatedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, totalCount, nil
}
