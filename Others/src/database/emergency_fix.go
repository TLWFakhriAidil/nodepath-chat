package database

import (
	"database/sql"
	"log"
)

// EmergencySequenceStepsFix runs immediately when called to fix the sequence steps issue
func EmergencySequenceStepsFix() {
	db := GetDB()
	if db == nil {
		log.Println("❌ Database not initialized, cannot run emergency fix")
		return
	}
	
	log.Println("🚨 RUNNING EMERGENCY SEQUENCE STEPS FIX...")
	
	// List of SQL commands to fix the issue
	fixes := []struct {
		name string
		sql  string
	}{
		{
			name: "Add missing trigger column",
			sql:  "ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS trigger VARCHAR(255);",
		},
		{
			name: "Add missing next_trigger column", 
			sql:  "ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS next_trigger VARCHAR(255);",
		},
		{
			name: "Add missing trigger_delay_hours column",
			sql:  "ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS trigger_delay_hours INTEGER DEFAULT 24;",
		},
		{
			name: "Add missing is_entry_point column",
			sql:  "ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS is_entry_point TINYINT(1) DEFAULT 0;",
		},
		{
			name: "Add missing image_url column",
			sql:  "ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS image_url TEXT;",
		},
		{
			name: "Add missing min_delay_seconds column",
			sql:  "ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 10;",
		},
		{
			name: "Add missing max_delay_seconds column",
			sql:  "ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 30;",
		},
		{
			name: "Add missing step_count to sequences",
			sql:  "ALTER TABLE sequences ADD COLUMN IF NOT EXISTS step_count INTEGER DEFAULT 0;",
		},
		{
			name: "Add missing total_steps to sequences",
			sql:  "ALTER TABLE sequences ADD COLUMN IF NOT EXISTS total_steps INTEGER DEFAULT 0;",
		},
		{
			name: "Fix NULL trigger values",
			sql:  `UPDATE sequence_steps SET trigger = CONCAT('step_', CAST(id AS CHAR)) 
				  WHERE trigger IS NULL OR trigger = '';`,
		},
		{
			name: "Fix NULL next_trigger values",
			sql: `UPDATE sequence_steps 
				  SET next_trigger = '' 
				  WHERE next_trigger IS NULL;`,
		},
		{
			name: "Fix NULL trigger_delay_hours values",
			sql: `UPDATE sequence_steps 
				  SET trigger_delay_hours = 24 
				  WHERE trigger_delay_hours IS NULL;`,
		},
		{
			name: "Fix NULL is_entry_point values",
			sql: `UPDATE sequence_steps 
				  SET is_entry_point = 0 
				  WHERE is_entry_point IS NULL;`,
		},
		{
			name: "Fix NULL min_delay_seconds values",
			sql: `UPDATE sequence_steps 
				  SET min_delay_seconds = 10 
				  WHERE min_delay_seconds IS NULL;`,
		},
		{
			name: "Fix NULL max_delay_seconds values",
			sql: `UPDATE sequence_steps 
				  SET max_delay_seconds = 30 
				  WHERE max_delay_seconds IS NULL;`,
		},
		{
			name: "Fix message_type values",
			sql: `UPDATE sequence_steps 
				  SET message_type = 'text' 
				  WHERE message_type IS NULL OR message_type = '';`,
		},
		{
			name: "Fix Invalid Date in send_time",
			sql: `UPDATE sequence_steps 
				  SET send_time = '10:00' 
				  WHERE send_time = 'Invalid Date' OR send_time IS NULL OR send_time = '';`,
		},
		{
			name: "Fix Invalid Date in time_schedule",
			sql: `UPDATE sequence_steps 
				  SET time_schedule = '10:00' 
				  WHERE time_schedule = 'Invalid Date' OR time_schedule IS NULL OR time_schedule = '';`,
		},
		{
			name: "Update sequence step counts",
			sql: `UPDATE sequences 
				  SET step_count = (SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = sequences.id),
				      total_steps = (SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = sequences.id);`,
		},
		{
			name: "Create performance index",
			sql: `CREATE INDEX IF NOT EXISTS idx_sequence_steps_complete ON sequence_steps(sequence_id, day_number);`,
		},
	}
	
	successCount := 0
	for _, fix := range fixes {
		_, err := db.Exec(fix.sql)
		if err != nil {
			log.Printf("⚠️  Fix '%s' failed: %v", fix.name, err)
		} else {
			log.Printf("✅ Fix '%s' completed", fix.name)
			successCount++
		}
	}
	
	log.Printf("🎉 Emergency fix completed! %d/%d fixes applied successfully", successCount, len(fixes))
	
	// Verify the fix worked
	verifyFix(db)
}

func verifyFix(db *sql.DB) {
	log.Println("🔍 Verifying sequence steps fix...")
	
	// Check if sequences now have step counts
	var sequenceCount, stepsCount int
	err := db.QueryRow("SELECT COUNT(*) FROM sequences WHERE step_count > 0").Scan(&sequenceCount)
	if err != nil {
		log.Printf("❌ Error checking sequences: %v", err)
		return
	}
	
	err = db.QueryRow("SELECT COUNT(*) from sequence_steps").Scan(&stepsCount)
	if err != nil {
		log.Printf("❌ Error checking steps: %v", err)
		return
	}
	
	log.Printf("📊 Results: %d sequences with steps, %d total steps in database", sequenceCount, stepsCount)
	
	if sequenceCount > 0 && stepsCount > 0 {
		log.Println("✅ Fix verification PASSED! Sequence steps should now work properly.")
	} else if stepsCount == 0 {
		log.Println("⚠️  No sequence steps found in database. Create some steps to test.")
	} else {
		log.Println("❌ Fix verification FAILED. Please check database manually.")
	}
}
