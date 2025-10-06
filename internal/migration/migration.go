package migration

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

// RunMigrations automatically runs all pending migrations
func (m *Migrator) RunMigrations() error {
	logrus.Info("Starting automatic database migration...")
	
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Apply critical migrations in order
	migrations := []struct {
		name string
		sql  string
	}{
		{
			name: "000012_add_user_profile_fields",
			sql: `-- Add missing columns to user_nodepath
ALTER TABLE user_nodepath 
ADD COLUMN IF NOT EXISTS gmail VARCHAR(255) NULL DEFAULT NULL COMMENT 'User Gmail address',
ADD COLUMN IF NOT EXISTS phone VARCHAR(20) NULL DEFAULT NULL COMMENT 'User phone number';

-- Update existing users to have active status if needed
UPDATE user_nodepath SET status = 'active' WHERE status IS NULL OR status = '';`,
		},
		{
			name: "000011_create_billing_nodepath",
			sql: `-- Create subscriptions_nodepath table
CREATE TABLE IF NOT EXISTS subscriptions_nodepath (
  id varchar(255) NOT NULL,
  user_id varchar(255) NOT NULL,
  plan_name varchar(255) NOT NULL DEFAULT 'Test Plan',
  plan_price decimal(10,2) NOT NULL DEFAULT 1.00,
  plan_period varchar(50) NOT NULL DEFAULT 'monthly',
  status enum('active','cancelled','suspended','pending') NOT NULL DEFAULT 'active',
  next_billing_date date NOT NULL,
  features json DEFAULT NULL,
  created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_user_id (user_id),
  KEY idx_status (status),
  KEY idx_next_billing_date (next_billing_date),
  KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create payments_nodepath table
CREATE TABLE IF NOT EXISTS payments_nodepath (
  id varchar(255) NOT NULL,
  user_id varchar(255) NOT NULL,
  subscription_id varchar(255) DEFAULT NULL,
  bill_id varchar(255) DEFAULT NULL COMMENT 'Billplz bill ID',
  invoice_number varchar(255) NOT NULL COMMENT 'Internal invoice number',
  amount decimal(10,2) NOT NULL,
  currency varchar(10) NOT NULL DEFAULT 'MYR',
  description text DEFAULT NULL,
  status enum('pending','paid','failed','cancelled') NOT NULL DEFAULT 'pending',
  payment_method varchar(50) DEFAULT 'billplz',
  billplz_url text DEFAULT NULL COMMENT 'Payment URL from Billplz',
  paid_at timestamp NULL DEFAULT NULL,
  created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY invoice_number (invoice_number),
  KEY idx_user_id (user_id),
  KEY idx_subscription_id (subscription_id),
  KEY idx_bill_id (bill_id),
  KEY idx_status (status),
  KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create billing_history_nodepath table
CREATE TABLE IF NOT EXISTS billing_history_nodepath (
  id varchar(255) NOT NULL,
  user_id varchar(255) NOT NULL,
  payment_id varchar(255) NOT NULL,
  invoice_number varchar(255) NOT NULL,
  amount decimal(10,2) NOT NULL,
  currency varchar(10) NOT NULL DEFAULT 'MYR',
  description text NOT NULL,
  status enum('pending','paid','failed','cancelled') NOT NULL,
  payment_date date DEFAULT NULL,
  created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_user_id (user_id),
  KEY idx_payment_id (payment_id),
  KEY idx_payment_date (payment_date),
  KEY idx_status (status),
  KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add foreign key constraints after tables are created
ALTER TABLE payments_nodepath 
ADD CONSTRAINT IF NOT EXISTS fk_payments_subscription 
FOREIGN KEY (subscription_id) REFERENCES subscriptions_nodepath (id) ON DELETE SET NULL;

ALTER TABLE billing_history_nodepath 
ADD CONSTRAINT IF NOT EXISTS fk_billing_payment 
FOREIGN KEY (payment_id) REFERENCES payments_nodepath (id) ON DELETE CASCADE;`,
		},
		{
			name: "insert_test_data",
			sql: `-- Insert test data for user ID 1 if it doesn't exist
INSERT IGNORE INTO subscriptions_nodepath 
(id, user_id, plan_name, plan_price, plan_period, status, next_billing_date, features) 
VALUES 
('test_sub_001', '1', 'Test Plan', 1.00, 'monthly', 'active', DATE_ADD(CURDATE(), INTERVAL 1 MONTH), 
JSON_ARRAY('WhatsApp Bot Integration', 'Flow Builder Access', 'Basic Analytics', 'Standard Support'));

INSERT IGNORE INTO payments_nodepath 
(id, user_id, subscription_id, invoice_number, amount, currency, description, status, payment_method) 
VALUES 
('test_pay_001', '1', 'test_sub_001', 'INV-TEST-001', 1.00, 'MYR', 'Test Plan - Monthly subscription', 'paid', 'billplz');

INSERT IGNORE INTO billing_history_nodepath 
(id, user_id, payment_id, invoice_number, amount, currency, description, status, payment_date) 
VALUES 
('test_hist_001', '1', 'test_pay_001', 'INV-TEST-001', 1.00, 'MYR', 'Test Plan - Monthly subscription', 'paid', CURDATE());`,
		},
	}

	// Run each migration
	for _, migration := range migrations {
		if err := m.runMigration(migration.name, migration.sql); err != nil {
			logrus.WithError(err).Errorf("Failed to run migration: %s", migration.name)
			return fmt.Errorf("migration failed: %s: %w", migration.name, err)
		}
		logrus.Infof("✅ Migration completed: %s", migration.name)
	}

	logrus.Info("🎉 All database migrations completed successfully!")
	return nil
}

// createMigrationsTable creates the migrations tracking table
func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	
	_, err := m.db.Exec(query)
	return err
}

// runMigration runs a single migration if it hasn't been applied yet
func (m *Migrator) runMigration(name, sql string) error {
	// Check if migration was already applied
	var count int
	err := m.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", name).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if count > 0 {
		logrus.Infof("⏭️  Migration already applied: %s", name)
		return nil
	}

	// Split SQL into individual statements
	statements := strings.Split(sql, ";")
	
	// Execute each statement
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		logrus.Debugf("Executing SQL: %s", stmt)
		if _, err := m.db.Exec(stmt); err != nil {
			// Log error but continue for some expected errors
			if strings.Contains(err.Error(), "Duplicate column name") ||
			   strings.Contains(err.Error(), "already exists") ||
			   strings.Contains(err.Error(), "Duplicate key name") {
				logrus.Warnf("⚠️  Expected error (continuing): %v", err)
				continue
			}
			return fmt.Errorf("failed to execute statement: %s: %w", stmt, err)
		}
	}

	// Mark migration as applied
	_, err = m.db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", name)
	if err != nil {
		return fmt.Errorf("failed to mark migration as applied: %w", err)
	}

	return nil
}

// LoadMigrationFiles loads migration files from the migrations directory
func (m *Migrator) LoadMigrationFiles(migrationsDir string) ([]string, error) {
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var upFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			upFiles = append(upFiles, file.Name())
		}
	}

	sort.Strings(upFiles)
	return upFiles, nil
}

// VerifyMigration checks if critical tables and columns exist
func (m *Migrator) VerifyMigration() error {
	logrus.Info("Verifying database migration...")

	// Check if user_nodepath has new columns
	columns := []string{"gmail", "phone", "status", "expired"}
	for _, column := range columns {
		var exists bool
		query := `
			SELECT COUNT(*) > 0 
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = 'user_nodepath' 
			AND COLUMN_NAME = ?
		`
		err := m.db.QueryRow(query, column).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check column %s: %w", column, err)
		}
		if !exists {
			logrus.Warnf("⚠️  Column missing: user_nodepath.%s", column)
		} else {
			logrus.Infof("✅ Column exists: user_nodepath.%s", column)
		}
	}

	// Check if billing tables exist
	tables := []string{"subscriptions_nodepath", "payments_nodepath", "billing_history_nodepath"}
	for _, table := range tables {
		var exists bool
		query := `
			SELECT COUNT(*) > 0 
			FROM INFORMATION_SCHEMA.TABLES 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = ?
		`
		err := m.db.QueryRow(query, table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}
		if !exists {
			logrus.Warnf("⚠️  Table missing: %s", table)
		} else {
			logrus.Infof("✅ Table exists: %s", table)
		}
	}

	// Check test data
	var count int
	err := m.db.QueryRow("SELECT COUNT(*) FROM subscriptions_nodepath WHERE id = 'test_sub_001'").Scan(&count)
	if err != nil {
		logrus.Warnf("⚠️  Could not verify test data: %v", err)
	} else if count > 0 {
		logrus.Info("✅ Test subscription data exists")
	} else {
		logrus.Warn("⚠️  Test subscription data missing")
	}

	logrus.Info("✅ Database migration verification completed")
	return nil
}