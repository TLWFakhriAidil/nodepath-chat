package utils

import (
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

// TransactionFunc represents a function that executes within a database transaction
type TransactionFunc func(tx *sql.Tx) error

// WithTransaction executes a function within a database transaction
// It automatically handles BEGIN, COMMIT, and ROLLBACK operations
// If the function returns an error, the transaction is rolled back
// If the function succeeds, the transaction is committed
func WithTransaction(db *sql.DB, fn TransactionFunc) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		logrus.WithError(err).Error("Failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is properly closed
	defer func() {
		if r := recover(); r != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logrus.WithError(rollbackErr).Error("Failed to rollback transaction after panic")
			}
			logrus.WithField("panic", r).Error("Transaction rolled back due to panic")
			panic(r) // Re-panic after cleanup
		}
	}()

	// Execute the function within the transaction
	err = fn(tx)
	if err != nil {
		// Rollback on error
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logrus.WithError(rollbackErr).Error("Failed to rollback transaction")
			return fmt.Errorf("transaction failed and rollback failed: %w (original error: %v)", rollbackErr, err)
		}
		logrus.WithError(err).Debug("Transaction rolled back due to error")
		return err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		logrus.WithError(err).Error("Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logrus.Debug("Transaction committed successfully")
	return nil
}

// WithTransactionRetry executes a function within a database transaction with retry logic
// It will retry the transaction up to maxRetries times if it fails due to deadlocks or timeouts
func WithTransactionRetry(db *sql.DB, fn TransactionFunc, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := WithTransaction(db, fn)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if this is a retryable error (deadlock, timeout, etc.)
		if !isRetryableError(err) {
			return err // Non-retryable error, fail immediately
		}

		if attempt < maxRetries {
			logrus.WithFields(logrus.Fields{
				"attempt":     attempt + 1,
				"max_retries": maxRetries,
				"error":       err.Error(),
			}).Warn("Transaction failed, retrying...")
		}
	}

	return fmt.Errorf("transaction failed after %d retries: %w", maxRetries, lastErr)
}

// isRetryableError checks if an error is retryable (deadlock, timeout, etc.)
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()

	// MySQL deadlock and timeout errors
	return contains(errorStr, "Deadlock found") ||
		contains(errorStr, "Lock wait timeout") ||
		contains(errorStr, "connection reset") ||
		contains(errorStr, "connection refused")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOfSubstring(s, substr) >= 0))
}

// indexOfSubstring finds the index of a substring in a string
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
