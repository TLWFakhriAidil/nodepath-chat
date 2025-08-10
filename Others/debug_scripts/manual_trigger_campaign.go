package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "time"
    
    _ "github.com/lib/pq"
)

func main() {
    // Get database URL from environment or use default
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        log.Fatal("DATABASE_URL environment variable is required")
    }
    
    // Connect to database
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // Test connection
    err = db.Ping()
    if err != nil {
        log.Fatal("Failed to ping database:", err)
    }
    
    fmt.Println("Connected to database successfully")
    
    // Check current time
    var serverTime time.Time
    err = db.QueryRow("SELECT NOW()").Scan(&serverTime)
    if err != nil {
        log.Fatal("Failed to get server time:", err)
    }
    
    fmt.Printf("Server time: %s\n", serverTime.Format("2006-01-02 15:04:05"))
    
    // Get campaign details
    var campaignID int
    var title, status, campaignDate, scheduledTime string
    var niche, targetStatus sql.NullString
    
    err = db.QueryRow(`
        SELECT id, title, status, campaign_date, 
               COALESCE(scheduled_time::text, '00:00:00'), 
               niche, COALESCE(target_status, 'all')
        FROM campaigns 
        WHERE title = 'test' 
        ORDER BY created_at DESC 
        LIMIT 1
    `).Scan(&campaignID, &title, &status, &campaignDate, &scheduledTime, &niche, &targetStatus)
    
    if err != nil {
        log.Fatal("Failed to get campaign:", err)
    }
    
    fmt.Printf("\nCampaign found:\n")
    fmt.Printf("ID: %d\n", campaignID)
    fmt.Printf("Title: %s\n", title)
    fmt.Printf("Status: %s\n", status)
    fmt.Printf("Date: %s\n", campaignDate)
    fmt.Printf("Time: %s\n", scheduledTime)
    fmt.Printf("Niche: %s\n", niche.String)
    fmt.Printf("Target Status: %s\n", targetStatus.String)
    
    if status == "sent" {
        fmt.Println("\n⚠️  Campaign already sent!")
        return
    }
    
    // Check for matching leads
    var leadCount int
    query := `
        SELECT COUNT(*) 
        FROM leads 
        WHERE niche LIKE '%' || $1 || '%'
        AND ($2 = 'all' OR target_status = $2)
    `
    err = db.QueryRow(query, niche.String, targetStatus.String).Scan(&leadCount)
    if err != nil {
        log.Fatal("Failed to count leads:", err)
    }
    
    fmt.Printf("\n📊 Found %d matching leads\n", leadCount)
    
    if leadCount == 0 {
        fmt.Println("❌ No leads match the campaign criteria!")
        return
    }
    
    // Ask for confirmation
    fmt.Printf("\n🚀 Ready to trigger campaign '%s' for %d leads?\n", title, leadCount)
    fmt.Print("Type 'yes' to continue: ")
    
    var confirm string
    fmt.Scanln(&confirm)
    
    if confirm != "yes" {
        fmt.Println("❌ Campaign trigger cancelled")
        return
    }
    
    // Update campaign status to trigger it
    _, err = db.Exec(`
        UPDATE campaigns 
        SET status = 'pending',
            updated_at = NOW()
        WHERE id = $1
    `, campaignID)
    
    if err != nil {
        log.Fatal("Failed to update campaign:", err)
    }
    
    fmt.Println("\n✅ Campaign status updated to 'pending'")
    fmt.Println("📨 The campaign trigger service will pick it up in the next minute")
    fmt.Println("👀 Check the Worker Status page to monitor progress")
}