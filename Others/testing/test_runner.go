package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// TestRunner simulates the entire system without sending real messages
type TestRunner struct {
	db              *sql.DB
	totalMessages   int64
	successMessages int64
	failedMessages  int64
	startTime       time.Time
	mu              sync.Mutex
	
	// Simulation parameters
	simulateDelay   bool
	minDelay        time.Duration
	maxDelay        time.Duration
	failureRate     float32
}

// MessageStats tracks message statistics
type MessageStats struct {
	DeviceID        string
	MessagesPerHour int
	TotalSent       int
	Failed          int
}

func main() {
	// Initialize logger
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://postgres:postgres@localhost/whatsapp?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create test runner
	runner := &TestRunner{
		db:            db,
		simulateDelay: true,
		minDelay:      5 * time.Second,
		maxDelay:      15 * time.Second,
		failureRate:   0.02, // 2% failure rate
		startTime:     time.Now(),
	}

	// Run test menu
	for {
		fmt.Println("\n========================================")
		fmt.Println("WhatsApp Multi-Device Test Runner")
		fmt.Println("========================================")
		fmt.Println("1. Generate Test Data (3000 devices, 50k leads)")
		fmt.Println("2. Simulate Campaign Broadcasting")
		fmt.Println("3. Simulate Sequence Processing") 
		fmt.Println("4. Simulate AI Campaign Distribution")
		fmt.Println("5. Run Full System Test (All features)")
		fmt.Println("6. Show Current Statistics")
		fmt.Println("7. Clean Test Data")
		fmt.Println("0. Exit")
		fmt.Print("\nSelect option: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			runner.generateTestData()
		case 2:
			runner.simulateCampaign()
		case 3:
			runner.simulateSequences()
		case 4:
			runner.simulateAICampaign()
		case 5:
			runner.runFullTest()
		case 6:
			runner.showStatistics()
		case 7:
			runner.cleanTestData()
		case 0:
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid option")
		}
	}
}

// generateTestData runs the SQL script to generate test data
func (r *TestRunner) generateTestData() {
	fmt.Println("\n🔄 Generating test data...")
	
	// Read SQL file
	sqlBytes, err := os.ReadFile("generate_test_data.sql")
	if err != nil {
		log.Printf("Error reading SQL file: %v", err)
		return
	}

	// Execute SQL
	_, err = r.db.Exec(string(sqlBytes))
	if err != nil {
		log.Printf("Error executing SQL: %v", err)
		return
	}

	fmt.Println("✅ Test data generated successfully!")
}

// simulateCampaign simulates campaign broadcasting
func (r *TestRunner) simulateCampaign() {
	fmt.Println("\n🚀 Starting Campaign Simulation...")
	
	// Get active campaigns
	rows, err := r.db.Query(`
		SELECT c.id, c.name, c.message, COUNT(l.id) as lead_count
		FROM campaigns c
		JOIN leads l ON l.user_id = c.user_id
		WHERE c.status = 'active' 
			AND c.user_id = (SELECT id FROM users WHERE email = 'test@whatsapp.com')
		GROUP BY c.id, c.name, c.message
	`)
	if err != nil {
		log.Printf("Error getting campaigns: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var campaignID, name, message string
		var leadCount int
		
		err := rows.Scan(&campaignID, &name, &message, &leadCount)
		if err != nil {
			continue
		}

		fmt.Printf("\n📢 Campaign: %s\n", name)
		fmt.Printf("   Leads to process: %d\n", leadCount)
		
		// Simulate sending to leads
		r.simulateBroadcast(campaignID, leadCount)
	}
}

// simulateBroadcast simulates sending messages
func (r *TestRunner) simulateBroadcast(campaignID string, totalLeads int) {
	// Get online devices
	deviceRows, err := r.db.Query(`
		SELECT id, device_name 
		FROM user_devices 
		WHERE status = 'online' 
			AND user_id = (SELECT id FROM users WHERE email = 'test@whatsapp.com')
		LIMIT 100
	`)
	if err != nil {
		return
	}
	defer deviceRows.Close()

	var devices []struct {
		ID   string
		Name string
	}
	
	for deviceRows.Next() {
		var d struct {
			ID   string
			Name string
		}
		deviceRows.Scan(&d.ID, &d.Name)
		devices = append(devices, d)
	}

	if len(devices) == 0 {
		fmt.Println("❌ No online devices found")
		return
	}

	fmt.Printf("   Using %d devices\n", len(devices))
	
	// Progress tracking
	var processed int64
	var wg sync.WaitGroup
	progressTicker := time.NewTicker(1 * time.Second)
	defer progressTicker.Stop()

	// Progress reporter
	go func() {
		for range progressTicker.C {
			p := atomic.LoadInt64(&processed)
			if p > 0 {
				elapsed := time.Since(r.startTime)
				rate := float64(p) / elapsed.Seconds()
				fmt.Printf("\r   Progress: %d/%d (%.0f msg/sec)", p, totalLeads, rate)
			}
		}
	}()

	// Simulate sending with multiple workers
	workers := 50
	leadChan := make(chan int, totalLeads)
	
	// Fill channel with lead indices
	for i := 0; i < totalLeads; i++ {
		leadChan <- i
	}
	close(leadChan)

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for range leadChan {
				// Pick random device (just for simulation, not actually used)
				_ = devices[rand.Intn(len(devices))]
				
				// Simulate delay
				if r.simulateDelay {
					delay := r.minDelay + time.Duration(rand.Int63n(int64(r.maxDelay-r.minDelay)))
					time.Sleep(delay / 100) // Speed up for simulation
				}
				
				// Simulate send
				if rand.Float32() > r.failureRate {
					atomic.AddInt64(&r.successMessages, 1)
				} else {
					atomic.AddInt64(&r.failedMessages, 1)
				}
				
				atomic.AddInt64(&processed, 1)
				atomic.AddInt64(&r.totalMessages, 1)
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("\n✅ Campaign simulation complete!\n")
}

// simulateSequences simulates sequence processing
func (r *TestRunner) simulateSequences() {
	fmt.Println("\n🔄 Starting Sequence Simulation...")
	
	// Get active sequences
	rows, err := r.db.Query(`
		SELECT s.id, s.name, s.trigger, COUNT(l.id) as matching_leads
		FROM sequences s
		JOIN leads l ON l.trigger = s.trigger AND l.user_id = s.user_id
		WHERE s.status = 'active'
			AND s.user_id = (SELECT id FROM users WHERE email = 'test@whatsapp.com')
		GROUP BY s.id, s.name, s.trigger
	`)
	if err != nil {
		log.Printf("Error getting sequences: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var seqID, name, trigger string
		var leadCount int
		
		err := rows.Scan(&seqID, &name, &trigger, &leadCount)
		if err != nil {
			continue
		}

		fmt.Printf("\n📋 Sequence: %s (Trigger: %s)\n", name, trigger)
		fmt.Printf("   Matching leads: %d\n", leadCount)
		fmt.Printf("   Simulating 30-day drip campaign...\n")
		
		// Simulate first day only for demo
		r.simulateBroadcast(seqID, leadCount/30) // Divide by 30 for daily simulation
	}
}

// simulateAICampaign simulates AI campaign distribution
func (r *TestRunner) simulateAICampaign() {
	fmt.Println("\n🤖 Starting AI Campaign Simulation...")
	
	// Simulate lead distribution to devices
	fmt.Println("   Distributing leads across devices...")
	time.Sleep(1 * time.Second)
	
	fmt.Println("   Applying smart throttling...")
	time.Sleep(1 * time.Second)
	
	fmt.Println("   Processing with human-like delays...")
	
	// Simulate some processing
	r.simulateBroadcast("ai-campaign", 5000)
}

// runFullTest runs all simulations
func (r *TestRunner) runFullTest() {
	fmt.Println("\n🎯 Running Full System Test...")
	fmt.Println("This will simulate all features with 3000 devices")
	
	r.startTime = time.Now()
	
	// Run all tests
	r.simulateCampaign()
	r.simulateSequences()
	r.simulateAICampaign()
	
	// Show final stats
	r.showStatistics()
}

// showStatistics displays current statistics
func (r *TestRunner) showStatistics() {
	fmt.Println("\n📊 System Statistics")
	fmt.Println("========================================")
	
	// Get device stats
	var totalDevices, onlineDevices int
	r.db.QueryRow(`
		SELECT COUNT(*), COUNT(CASE WHEN status = 'online' THEN 1 END)
		FROM user_devices
		WHERE device_name LIKE 'TestDevice%'
	`).Scan(&totalDevices, &onlineDevices)
	
	// Get lead stats
	var totalLeads int
	r.db.QueryRow(`
		SELECT COUNT(*) FROM leads WHERE name LIKE 'TestLead%'
	`).Scan(&totalLeads)
	
	elapsed := time.Since(r.startTime)
	
	fmt.Printf("Devices: %d total, %d online (%.1f%%)\n", 
		totalDevices, onlineDevices, float64(onlineDevices)/float64(totalDevices)*100)
	fmt.Printf("Leads: %d\n", totalLeads)
	fmt.Printf("\nMessages Processed:\n")
	fmt.Printf("  Total: %d\n", r.totalMessages)
	fmt.Printf("  Success: %d (%.1f%%)\n", 
		r.successMessages, float64(r.successMessages)/float64(r.totalMessages)*100)
	fmt.Printf("  Failed: %d (%.1f%%)\n",
		r.failedMessages, float64(r.failedMessages)/float64(r.totalMessages)*100)
	
	if elapsed > 0 {
		rate := float64(r.totalMessages) / elapsed.Seconds()
		fmt.Printf("\nPerformance:\n")
		fmt.Printf("  Runtime: %v\n", elapsed.Round(time.Second))
		fmt.Printf("  Rate: %.0f messages/second\n", rate)
		fmt.Printf("  Projected hourly: %.0f messages\n", rate*3600)
	}
}

// cleanTestData removes all test data
func (r *TestRunner) cleanTestData() {
	fmt.Print("\n⚠️  This will delete all test data. Are you sure? (y/n): ")
	var confirm string
	fmt.Scanln(&confirm)
	
	if confirm != "y" {
		return
	}
	
	fmt.Println("🧹 Cleaning test data...")
	
	queries := []string{
		"DELETE FROM broadcast_messages WHERE device_id IN (SELECT id FROM user_devices WHERE device_name LIKE 'TestDevice%')",
		"DELETE FROM sequence_contacts WHERE sequence_id IN (SELECT id FROM sequences WHERE name LIKE 'Test Sequence%')",
		"DELETE FROM sequence_steps WHERE sequence_id IN (SELECT id FROM sequences WHERE name LIKE 'Test Sequence%')",
		"DELETE FROM sequences WHERE name LIKE 'Test Sequence%'",
		"DELETE FROM ai_campaigns WHERE campaign_name LIKE 'Test AI Campaign%'",
		"DELETE FROM campaigns WHERE name LIKE 'Test Campaign%'",
		"DELETE FROM leads WHERE name LIKE 'TestLead%'",
		"DELETE FROM user_devices WHERE device_name LIKE 'TestDevice%'",
		"DELETE FROM users WHERE email = 'test@whatsapp.com'",
	}
	
	for _, query := range queries {
		_, err := r.db.Exec(query)
		if err != nil {
			log.Printf("Error executing cleanup: %v", err)
		}
	}
	
	fmt.Println("✅ Test data cleaned successfully!")
}
