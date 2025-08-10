package main

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/lib/pq"
)

func main() {
    // Connect to database
    db, err := sql.Open("postgres", "postgresql://whatsappusecase:4Lf!n7kB9pQ2sXw@roundhouse.proxy.rlwy.net:42790/railway")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Check device status
    deviceID := "d409cadc-75e2-4004-a789-c2bad0b31393"
    
    var status, platform sql.NullString
    err = db.QueryRow("SELECT status, platform FROM user_devices WHERE id = $1", deviceID).Scan(&status, &platform)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Device %s:\n", deviceID)
    fmt.Printf("Status: %s\n", status.String)
    fmt.Printf("Platform: %s\n", platform.String)
    
    // Check campaign messages
    var pendingCount, queuedCount, skippedCount int
    db.QueryRow("SELECT COUNT(*) FROM broadcast_messages WHERE device_id = $1 AND status = 'pending'", deviceID).Scan(&pendingCount)
    db.QueryRow("SELECT COUNT(*) FROM broadcast_messages WHERE device_id = $1 AND status = 'queued'", deviceID).Scan(&queuedCount)
    db.QueryRow("SELECT COUNT(*) FROM broadcast_messages WHERE device_id = $1 AND status = 'skipped'", deviceID).Scan(&skippedCount)
    
    fmt.Printf("\nBroadcast Messages:\n")
    fmt.Printf("Pending: %d\n", pendingCount)
    fmt.Printf("Queued: %d\n", queuedCount)
    fmt.Printf("Skipped: %d\n", skippedCount)
    
    // Show recent messages
    fmt.Printf("\nRecent broadcast messages:\n")
    rows, _ := db.Query(`
        SELECT id, campaign_id, status, error_message, created_at 
        FROM broadcast_messages 
        WHERE device_id = $1 
        ORDER BY created_at DESC 
        LIMIT 5
    `, deviceID)
    defer rows.Close()
    
    for rows.Next() {
        var id, status string
        var campaignID sql.NullInt64
        var errorMsg sql.NullString
        var createdAt string
        
        rows.Scan(&id, &campaignID, &status, &errorMsg, &createdAt)
        fmt.Printf("- ID: %s, Campaign: %d, Status: %s, Error: %s\n", 
            id[:8], campaignID.Int64, status, errorMsg.String)
    }
}
