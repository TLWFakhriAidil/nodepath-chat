package main

import (
	"database/sql"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"nodepath-chat/internal/config"
	"nodepath-chat/internal/database"
	"nodepath-chat/internal/migration"
)

func main() {
	logrus.Info("NodePath Chat - Database Migration Tool")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Println("No .env file found, using environment variables")
	}

	// Check if MYSQL_URI is set
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		logrus.Fatal("MYSQL_URI environment variable is required")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize database")
	}
	defer db.Close()

	logrus.Info("Database connection established")

	// Create migrator
	migrator := migration.NewMigrator(db)

	// Run migrations
	logrus.Info("Starting database migration...")
	if err := migrator.RunMigrations(); err != nil {
		logrus.WithError(err).Fatal("Migration failed")
	}

	// Verify migration
	logrus.Info("Verifying migration...")
	if err := migrator.VerifyMigration(); err != nil {
		logrus.WithError(err).Error("Migration verification failed")
	}

	// Show final status
	showDatabaseStatus(db)

	logrus.Info("🎉 Migration completed successfully!")
}

func showDatabaseStatus(db *sql.DB) {
	logrus.Info("=== DATABASE STATUS ===")

	// Check user_nodepath columns
	var columns []string
	rows, err := db.Query(`
		SELECT COLUMN_NAME 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'user_nodepath'
		ORDER BY ORDINAL_POSITION
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var col string
			if err := rows.Scan(&col); err == nil {
				columns = append(columns, col)
			}
		}
		logrus.Infof("user_nodepath columns: %v", columns)
	}

	// Check billing tables
	var tables []string
	rows, err = db.Query(`
		SELECT TABLE_NAME 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME LIKE '%_nodepath'
		ORDER BY TABLE_NAME
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var table string
			if err := rows.Scan(&table); err == nil {
				tables = append(tables, table)
			}
		}
		logrus.Infof("Tables with _nodepath suffix: %v", tables)
	}

	// Check test data
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM subscriptions_nodepath").Scan(&count)
	if err == nil {
		logrus.Infof("Subscription records: %d", count)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM payments_nodepath").Scan(&count)
	if err == nil {
		logrus.Infof("Payment records: %d", count)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM billing_history_nodepath").Scan(&count)
	if err == nil {
		logrus.Infof("Billing history records: %d", count)
	}

	logrus.Info("======================")
}