package main

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    // Connect to database
    db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/waweb")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Test if processing_worker_id column exists
    row := db.QueryRow("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = 'waweb' AND table_name = 'broadcast_messages' AND column_name = 'processing_worker_id'")
    var count int
    err = row.Scan(&count)
    if err != nil {
        log.Fatal(err)
    }
    
    if count == 0 {
        fmt.Println("ERROR: processing_worker_id column does NOT exist!")
    } else {
        fmt.Println("OK: processing_worker_id column exists")
    }

    // Check current messages
    rows, err := db.Query(`
        SELECT id, status, processing_worker_id, processing_started_at 
        FROM broadcast_messages 
        WHERE status IN ('pending', 'processing', 'queued') 
        LIMIT 5
    `)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    fmt.Println("\nCurrent messages:")
    for rows.Next() {
        var id, status string
        var workerID, startedAt sql.NullString
        err := rows.Scan(&id, &status, &workerID, &startedAt)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("ID: %s, Status: %s, WorkerID: %v, StartedAt: %v\n", 
            id, status, workerID.String, startedAt.String)
    }

    // Test update
    testID := "test-worker-123"
    result, err := db.Exec(`
        UPDATE broadcast_messages 
        SET processing_worker_id = ?
        WHERE id = (SELECT id FROM (SELECT id FROM broadcast_messages WHERE status = 'pending' LIMIT 1) AS temp)
    `, testID)
    
    if err != nil {
        fmt.Printf("\nERROR updating: %v\n", err)
    } else {
        rows, _ := result.RowsAffected()
        fmt.Printf("\nUpdated %d rows with test worker ID\n", rows)
    }
}
