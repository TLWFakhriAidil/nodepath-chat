package main

import (
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("=== CHECKING DATABASE TABLES ===")

	// Connect to database
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local"
	}

	db, err := gorm.Open(mysql.Open(mysqlURI), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ Error connecting to database: %v\n", err)
		return
	}
	fmt.Println("✅ Database connected successfully")

	// Get all tables with 'nodepath' in the name
	var tables []string
	err = db.Raw("SHOW TABLES LIKE '%nodepath%'").Scan(&tables).Error
	if err != nil {
		fmt.Printf("❌ Error getting tables: %v\n", err)
		return
	}

	fmt.Printf("\n📋 Found %d tables with 'nodepath':\n", len(tables))
	for i, table := range tables {
		fmt.Printf("   %d. %s\n", i+1, table)
	}

	// Check each table structure
	for _, table := range tables {
		fmt.Printf("\n🔍 Table: %s\n", table)
		fmt.Println("=" + fmt.Sprintf("%*s", len(table)+8, ""))
		
		type ColumnInfo struct {
			Field   string `gorm:"column:Field"`
			Type    string `gorm:"column:Type"`
			Null    string `gorm:"column:Null"`
			Key     string `gorm:"column:Key"`
			Default string `gorm:"column:Default"`
			Extra   string `gorm:"column:Extra"`
		}
		
		var columns []ColumnInfo
		err = db.Raw("DESCRIBE " + table).Scan(&columns).Error
		if err != nil {
			fmt.Printf("❌ Error describing table %s: %v\n", table, err)
			continue
		}
		
		fmt.Printf("Columns (%d):\n", len(columns))
		for _, col := range columns {
			fmt.Printf("  - %s (%s)\n", col.Field, col.Type)
		}
		
		// Check if this looks like the conversation table
		hasConvLast := false
		hasConvCurrent := false
		hasProspectNum := false
		hasStage := false
		
		for _, col := range columns {
			switch col.Field {
			case "conv_last":
				hasConvLast = true
			case "conv_current":
				hasConvCurrent = true
			case "prospect_num":
				hasProspectNum = true
			case "stage":
				hasStage = true
			}
		}
		
		if hasConvLast && hasConvCurrent && hasProspectNum && hasStage {
			fmt.Printf("🎯 THIS IS LIKELY THE CONVERSATION TABLE!\n")
			
			// Check if there are any records
			var count int64
			err = db.Raw("SELECT COUNT(*) FROM " + table).Scan(&count).Error
			if err != nil {
				fmt.Printf("❌ Error counting records: %v\n", err)
			} else {
				fmt.Printf("📊 Record count: %d\n", count)
				
				if count > 0 {
					// Show sample records
					type SampleRecord struct {
						IDDevice    string `gorm:"column:id_device"`
						ProspectNum string `gorm:"column:prospect_num"`
						Stage       string `gorm:"column:stage"`
						ConvLast    string `gorm:"column:conv_last"`
						ConvCurrent string `gorm:"column:conv_current"`
					}
					
					var samples []SampleRecord
					err = db.Raw("SELECT id_device, prospect_num, stage, conv_last, conv_current FROM " + table + " LIMIT 5").Scan(&samples).Error
					if err != nil {
						fmt.Printf("❌ Error getting sample records: %v\n", err)
					} else {
						fmt.Printf("\n📝 Sample records:\n")
						for i, sample := range samples {
							fmt.Printf("   %d. Device: %s, Prospect: %s, Stage: %s\n", i+1, sample.IDDevice, sample.ProspectNum, sample.Stage)
							fmt.Printf("      Conv Last: %s\n", getPreview(sample.ConvLast, 100))
							fmt.Printf("      Conv Current: %s\n", getPreview(sample.ConvCurrent, 100))
							fmt.Println()
						}
					}
				}
			}
		}
	}

	fmt.Println("\n=== TABLE CHECK COMPLETE ===")
}

func getPreview(text string, maxLen int) string {
	if len(text) == 0 {
		return "[EMPTY]"
	}
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}