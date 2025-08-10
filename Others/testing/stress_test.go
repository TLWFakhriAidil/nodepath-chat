package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
)

// StressTest performs various stress testing scenarios
type StressTest struct {
	db *sql.DB
}

func main() {
	// Connect to database
	db, err := sql.Open("postgres", "postgresql://postgres:postgres@localhost/whatsapp?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	st := &StressTest{db: db}

	fmt.Println("\n🔥 WhatsApp Stress Testing Suite")
	fmt.Println("================================")

	for {
		fmt.Println("\n1. Device Churn Test (Rapid connect/disconnect)")
		fmt.Println("2. Message Burst Test (Sudden high load)")
		fmt.Println("3. Database Stress Test (Heavy concurrent queries)")
		fmt.Println("4. Memory Leak Test (Long running simulation)")
		fmt.Println("5. Failover Test (Mass device failures)")
		fmt.Println("6. Rate Limit Test (Test throttling)")
		fmt.Println("0. Exit")
		fmt.Print("\nSelect test: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			st.deviceChurnTest()
		case 2:
			st.messageBurstTest()
		case 3:
			st.databaseStressTest()
		case 4:
			st.memoryLeakTest()
		case 5:
			st.failoverTest()
		case 6:
			st.rateLimitTest()
		case 0:
			return
		}
	}
}
// deviceChurnTest simulates devices rapidly connecting and disconnecting
func (st *StressTest) deviceChurnTest() {
	fmt.Println("\n🔄 Starting Device Churn Test...")
	fmt.Println("Simulating rapid connect/disconnect cycles")

	duration := 30 * time.Second
	endTime := time.Now().Add(duration)
	cycles := int64(0)
	errors := int64(0)

	var wg sync.WaitGroup
	deviceCount := 100 // Test with 100 devices

	for i := 0; i < deviceCount; i++ {
		wg.Add(1)
		go func(deviceNum int) {
			defer wg.Done()
			deviceID := fmt.Sprintf("TestDevice%04d", deviceNum)

			for time.Now().Before(endTime) {
				// Toggle online/offline
				status := "offline"
				if atomic.LoadInt64(&cycles)%2 == 0 {
					status = "online"
				}

				_, err := st.db.Exec(`
					UPDATE user_devices 
					SET status = $1, last_seen = NOW() 
					WHERE device_name = $2
				`, status, deviceID)

				if err != nil {
					atomic.AddInt64(&errors, 1)
				}
				atomic.AddInt64(&cycles, 1)

				time.Sleep(100 * time.Millisecond)
			}
		}(i)
	}

	// Monitor progress
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			if time.Now().After(endTime) {
				break
			}
			c := atomic.LoadInt64(&cycles)
			e := atomic.LoadInt64(&errors)
			fmt.Printf("\rCycles: %d, Errors: %d, Rate: %.0f/sec", c, e, float64(c)/time.Since(endTime.Add(-duration)).Seconds())
		}
	}()

	wg.Wait()
	fmt.Printf("\n✅ Churn test complete: %d cycles, %d errors\n", cycles, errors)
}

// messageBurstTest simulates sudden spike in message sending
func (st *StressTest) messageBurstTest() {
	fmt.Println("\n💥 Starting Message Burst Test...")
	fmt.Println("Simulating sudden 10x load increase")

	// Normal load for 10 seconds
	fmt.Println("Phase 1: Normal load (100 msg/sec)")
	st.simulateLoad(100, 10*time.Second)

	// Sudden burst
	fmt.Println("Phase 2: BURST! (1000 msg/sec)")
	st.simulateLoad(1000, 10*time.Second)

	// Return to normal
	fmt.Println("Phase 3: Recovery (100 msg/sec)")
	st.simulateLoad(100, 10*time.Second)

	fmt.Println("✅ Burst test complete")
}

// simulateLoad generates specified message load
func (st *StressTest) simulateLoad(messagesPerSec int, duration time.Duration) {
	sent := int64(0)
	failed := int64(0)
	endTime := time.Now().Add(duration)

	var wg sync.WaitGroup
	workers := 10
	messagesPerWorker := messagesPerSec / workers

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(time.Second / time.Duration(messagesPerWorker))
			defer ticker.Stop()

			for time.Now().Before(endTime) {
				select {
				case <-ticker.C:
					// Simulate message insert
					_, err := st.db.Exec(`
						INSERT INTO broadcast_messages (id, campaign_id, lead_id, device_id, status, created_at)
						VALUES (gen_random_uuid(), 
							(SELECT id FROM campaigns LIMIT 1),
							(SELECT id FROM leads LIMIT 1),
							(SELECT id FROM user_devices WHERE status = 'online' LIMIT 1),
							'sent', NOW())
					`)
					if err != nil {
						atomic.AddInt64(&failed, 1)
					} else {
						atomic.AddInt64(&sent, 1)
					}
				}
			}
		}()
	}

	// Monitor
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		start := time.Now()
		
		for range ticker.C {
			if time.Now().After(endTime) {
				break
			}
			s := atomic.LoadInt64(&sent)
			f := atomic.LoadInt64(&failed)
			rate := float64(s) / time.Since(start).Seconds()
			fmt.Printf("\rSent: %d, Failed: %d, Rate: %.0f/sec", s, f, rate)
		}
	}()

	wg.Wait()
	fmt.Printf("\n")
}

// databaseStressTest tests database under heavy concurrent load
func (st *StressTest) databaseStressTest() {
	fmt.Println("\n🗄️ Starting Database Stress Test...")
	fmt.Println("Testing with 200 concurrent connections")

	queries := []string{
		// Heavy read query
		`SELECT COUNT(*) FROM leads WHERE status = 'Active'`,
		// Join query
		`SELECT COUNT(*) FROM broadcast_messages bm 
		 JOIN leads l ON l.id = bm.lead_id 
		 WHERE bm.created_at > NOW() - INTERVAL '1 hour'`,
		// Update query
		`UPDATE user_devices SET last_seen = NOW() WHERE id = (SELECT id FROM user_devices LIMIT 1)`,
		// Insert query
		`INSERT INTO leads (id, user_id, device_id, name, phone, status, created_at, updated_at)
		 VALUES (gen_random_uuid(), 
			(SELECT id FROM users LIMIT 1),
			(SELECT id FROM user_devices LIMIT 1),
			'StressTest', '60123456789', 'Active', NOW(), NOW())
		 ON CONFLICT DO NOTHING`,
	}

	var wg sync.WaitGroup
	completed := int64(0)
	errors := int64(0)
	duration := 30 * time.Second
	endTime := time.Now().Add(duration)

	// Launch concurrent workers
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for time.Now().Before(endTime) {
				query := queries[workerID%len(queries)]
				_, err := st.db.Exec(query)
				
				if err != nil {
					atomic.AddInt64(&errors, 1)
				} else {
					atomic.AddInt64(&completed, 1)
				}
			}
		}(i)
	}

	// Monitor
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			c := atomic.LoadInt64(&completed)
			e := atomic.LoadInt64(&errors)
			fmt.Printf("\rQueries: %d, Errors: %d, QPS: %.0f", c, e, float64(c)/time.Since(endTime.Add(-duration)).Seconds())
		}
	}()

	wg.Wait()
	ticker.Stop()
	fmt.Printf("\n✅ Database stress test complete\n")
}

// memoryLeakTest runs continuous operations to check for memory leaks
func (st *StressTest) memoryLeakTest() {
	fmt.Println("\n🧠 Starting Memory Leak Test...")
	fmt.Println("Running for 2 minutes - monitor memory usage")
	fmt.Println("Press Ctrl+C to stop early")

	duration := 2 * time.Minute
	endTime := time.Now().Add(duration)
	operations := int64(0)

	// Continuous operations
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for time.Now().Before(endTime) {
				// Create and discard data
				var count int
				st.db.QueryRow("SELECT COUNT(*) FROM leads").Scan(&count)
				
				// Simulate object creation
				data := make([]byte, 1024*10) // 10KB
				_ = data
				
				atomic.AddInt64(&operations, 1)
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	// Progress monitor
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			if time.Now().After(endTime) {
				break
			}
			ops := atomic.LoadInt64(&operations)
			remaining := endTime.Sub(time.Now()).Round(time.Second)
			fmt.Printf("\rOperations: %d, Time remaining: %v", ops, remaining)
		}
	}()

	wg.Wait()
	ticker.Stop()
	fmt.Printf("\n✅ Memory leak test complete - check memory usage\n")
}

// failoverTest simulates mass device failures
func (st *StressTest) failoverTest() {
	fmt.Println("\n⚠️ Starting Failover Test...")
	fmt.Println("Simulating 50% device failure")

	// Mark 50% devices as offline
	result, err := st.db.Exec(`
		UPDATE user_devices 
		SET status = 'offline' 
		WHERE device_name LIKE 'TestDevice%' 
		AND device_name IN (
			SELECT device_name FROM user_devices 
			WHERE device_name LIKE 'TestDevice%' 
			ORDER BY RANDOM() 
			LIMIT (SELECT COUNT(*)/2 FROM user_devices WHERE device_name LIKE 'TestDevice%')
		)
	`)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	affected, _ := result.RowsAffected()
	fmt.Printf("Marked %d devices as offline\n", affected)

	// Test message sending with reduced capacity
	fmt.Println("Testing message sending with reduced capacity...")
	st.simulateLoad(200, 10*time.Second)

	// Restore devices
	fmt.Println("Restoring devices...")
	st.db.Exec(`UPDATE user_devices SET status = 'online' WHERE device_name LIKE 'TestDevice%'`)
	
	fmt.Println("✅ Failover test complete")
}

// rateLimitTest tests rate limiting behavior
func (st *StressTest) rateLimitTest() {
	fmt.Println("\n⏱️ Starting Rate Limit Test...")
	fmt.Println("Testing 80 msg/hour per device limit")

	// Get a test device
	var deviceID string
	err := st.db.QueryRow(`
		SELECT id FROM user_devices 
		WHERE device_name LIKE 'TestDevice%' 
		AND status = 'online' 
		LIMIT 1
	`).Scan(&deviceID)
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Try to send 100 messages rapidly
	sent := 0
	failed := 0
	
	for i := 0; i < 100; i++ {
		// Check current hour count
		var hourCount int
		st.db.QueryRow(`
			SELECT COUNT(*) FROM broadcast_messages 
			WHERE device_id = $1 
			AND created_at > NOW() - INTERVAL '1 hour'
		`, deviceID).Scan(&hourCount)

		if hourCount >= 80 {
			failed++
			fmt.Printf("\rRate limit hit at message %d (hourly: %d)", i+1, hourCount)
		} else {
			// Send message
			_, err := st.db.Exec(`
				INSERT INTO broadcast_messages (id, device_id, lead_id, status, created_at)
				VALUES (gen_random_uuid(), $1, 
					(SELECT id FROM leads LIMIT 1),
					'sent', NOW())
			`, deviceID)
			
			if err != nil {
				failed++
			} else {
				sent++
			}
		}
		
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\n✅ Rate limit test complete: %d sent, %d blocked\n", sent, failed)
}
