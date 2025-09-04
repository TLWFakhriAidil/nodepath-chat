package main

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    // Connect to database
    db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/nodepath_chat")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Get table structure
    rows, err := db.Query("DESCRIBE ai_whatsapp_nodepath")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    fmt.Println("\n=== ai_whatsapp_nodepath Table Structure ===")
    fmt.Printf("%-30s %-30s %-10s %-10s %-20s %s\n", "Field", "Type", "Null", "Key", "Default", "Extra")
    fmt.Println(string(make([]byte, 120)))

    for rows.Next() {
        var field, typ, null, key, extra string
        var def sql.NullString
        err := rows.Scan(&field, &typ, &null, &key, &def, &extra)
        if err != nil {
            log.Fatal(err)
        }
        defaultVal := "NULL"
        if def.Valid {
            defaultVal = def.String
        }
        fmt.Printf("%-30s %-30s %-10s %-10s %-20s %s\n", field, typ, null, key, defaultVal, extra)
    }
}