# AI WhatsApp Table Column Migration Script
# ==========================================
# This script helps fix the ai_whatsapp_nodepath table columns

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "AI WhatsApp Table Column Migration Tool" -ForegroundColor Yellow
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Check if MYSQL_URI is set
$mysqlUri = $env:MYSQL_URI
if (-not $mysqlUri) {
    Write-Host "❌ MYSQL_URI environment variable is not set!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please set it using:" -ForegroundColor Yellow
    Write-Host '$env:MYSQL_URI = "mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway"' -ForegroundColor Green
    Write-Host ""
    exit 1
}

Write-Host "✅ MYSQL_URI is set" -ForegroundColor Green
Write-Host ""

# Menu
Write-Host "Choose an option:" -ForegroundColor Yellow
Write-Host "1. Run Go migration tool (Recommended)" -ForegroundColor White
Write-Host "2. Generate SQL script for manual execution" -ForegroundColor White
Write-Host "3. Test database connection only" -ForegroundColor White
Write-Host "4. Show current table structure" -ForegroundColor White
Write-Host ""

$choice = Read-Host "Enter your choice (1-4)"

switch ($choice) {
    "1" {
        Write-Host ""
        Write-Host "🔧 Running Go migration tool..." -ForegroundColor Cyan
        Write-Host "================================" -ForegroundColor Cyan
        
        # Check if the Go file exists
        $goFile = "debug\fix_ai_whatsapp_columns.go"
        if (-not (Test-Path $goFile)) {
            Write-Host "❌ Migration file not found: $goFile" -ForegroundColor Red
            exit 1
        }
        
        # Run the Go migration
        go run $goFile
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host ""
            Write-Host "✅ Migration completed successfully!" -ForegroundColor Green
        } else {
            Write-Host ""
            Write-Host "❌ Migration failed with exit code: $LASTEXITCODE" -ForegroundColor Red
        }
    }
    
    "2" {
        Write-Host ""
        Write-Host "📋 SQL script has been generated: fix_ai_whatsapp_table.sql" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "You can run this script in your MySQL client:" -ForegroundColor Yellow
        Write-Host "1. Open phpMyAdmin or MySQL Workbench" -ForegroundColor White
        Write-Host "2. Connect to your database" -ForegroundColor White
        Write-Host "3. Run the script: fix_ai_whatsapp_table.sql" -ForegroundColor White
        Write-Host ""
        Write-Host "Or use MySQL command line:" -ForegroundColor Yellow
        Write-Host "mysql -h 157.245.206.124 -u admin_aqil -p admin_railway < fix_ai_whatsapp_table.sql" -ForegroundColor Green
    }
    
    "3" {
        Write-Host ""
        Write-Host "🔍 Testing database connection..." -ForegroundColor Cyan
        Write-Host "=================================" -ForegroundColor Cyan
        
        # Create a simple Go test file
        $testCode = @'
package main

import (
    "database/sql"
    "fmt"
    "os"
    "strings"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    mysqlURI := os.Getenv("MYSQL_URI")
    if mysqlURI == "" {
        fmt.Println("❌ MYSQL_URI not set")
        os.Exit(1)
    }
    
    if strings.HasPrefix(mysqlURI, "mysql://") {
        mysqlURI = convertMySQLURIToDSN(mysqlURI)
    }
    
    db, err := sql.Open("mysql", mysqlURI)
    if err != nil {
        fmt.Printf("❌ Failed to open database: %v\n", err)
        os.Exit(1)
    }
    defer db.Close()
    
    if err := db.Ping(); err != nil {
        fmt.Printf("❌ Failed to ping database: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("✅ Database connection successful!")
    
    // Check if table exists
    var tableName string
    err = db.QueryRow("SHOW TABLES LIKE 'ai_whatsapp_nodepath'").Scan(&tableName)
    if err != nil {
        fmt.Println("⚠️  Table ai_whatsapp_nodepath does not exist")
    } else {
        fmt.Println("✅ Table ai_whatsapp_nodepath exists")
        
        // Count columns
        var count int
        db.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath'").Scan(&count)
        fmt.Printf("📊 Table has %d columns\n", count)
    }
}

func convertMySQLURIToDSN(uri string) string {
    uri = strings.TrimPrefix(uri, "mysql://")
    parts := strings.Split(uri, "@")
    if len(parts) != 2 {
        return uri
    }
    userPass := parts[0]
    hostDB := parts[1]
    hostDB = strings.Replace(hostDB, "/", ")/", 1)
    return userPass + "@tcp(" + hostDB + "?parseTime=true&charset=utf8mb4"
}
'@
        
        $testCode | Out-File -FilePath "debug\test_connection.go" -Encoding UTF8
        go run debug\test_connection.go
        Remove-Item "debug\test_connection.go" -ErrorAction SilentlyContinue
    }
    
    "4" {
        Write-Host ""
        Write-Host "📊 Checking current table structure..." -ForegroundColor Cyan
        Write-Host "======================================" -ForegroundColor Cyan
        
        # Create a Go script to show table structure
        $showCode = @'
package main

import (
    "database/sql"
    "fmt"
    "os"
    "strings"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    mysqlURI := os.Getenv("MYSQL_URI")
    if mysqlURI == "" {
        fmt.Println("❌ MYSQL_URI not set")
        os.Exit(1)
    }
    
    if strings.HasPrefix(mysqlURI, "mysql://") {
        mysqlURI = convertMySQLURIToDSN(mysqlURI)
    }
    
    db, err := sql.Open("mysql", mysqlURI)
    if err != nil {
        fmt.Printf("❌ Failed to open database: %v\n", err)
        os.Exit(1)
    }
    defer db.Close()
    
    // Get columns
    rows, err := db.Query(`
        SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT
        FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = 'ai_whatsapp_nodepath'
        ORDER BY ORDINAL_POSITION
    `)
    if err != nil {
        fmt.Printf("❌ Failed to get columns: %v\n", err)
        os.Exit(1)
    }
    defer rows.Close()
    
    fmt.Println("\nCurrent columns in ai_whatsapp_nodepath:")
    fmt.Println("=========================================")
    
    count := 0
    for rows.Next() {
        var name, dataType, nullable, defaultVal, comment sql.NullString
        rows.Scan(&name, &dataType, &nullable, &defaultVal, &comment)
        
        fmt.Printf("%2d. %-20s %-20s %s\n", count+1, name.String, dataType.String, nullable.String)
        if comment.Valid && comment.String != "" {
            fmt.Printf("    Comment: %s\n", comment.String)
        }
        count++
    }
    
    fmt.Printf("\nTotal columns: %d\n", count)
    
    // Check for unwanted columns
    unwanted := []string{"jam", "catatan_staff", "data_image", "conv_stage", "variables", "bot_balas", "current_node"}
    fmt.Println("\nChecking for unwanted columns:")
    fmt.Println("==============================")
    
    for _, col := range unwanted {
        var exists int
        db.QueryRow(`
            SELECT COUNT(*) 
            FROM INFORMATION_SCHEMA.COLUMNS 
            WHERE TABLE_SCHEMA = DATABASE() 
            AND TABLE_NAME = 'ai_whatsapp_nodepath' 
            AND COLUMN_NAME = ?
        `, col).Scan(&exists)
        
        if exists > 0 {
            fmt.Printf("❌ %s - EXISTS (should be removed)\n", col)
        } else {
            fmt.Printf("✅ %s - not found (good)\n", col)
        }
    }
}

func convertMySQLURIToDSN(uri string) string {
    uri = strings.TrimPrefix(uri, "mysql://")
    parts := strings.Split(uri, "@")
    if len(parts) != 2 {
        return uri
    }
    userPass := parts[0]
    hostDB := parts[1]
    hostDB = strings.Replace(hostDB, "/", ")/", 1)
    return userPass + "@tcp(" + hostDB + "?parseTime=true&charset=utf8mb4"
}
'@
        
        $showCode | Out-File -FilePath "debug\show_structure.go" -Encoding UTF8
        go run debug\show_structure.go
        Remove-Item "debug\show_structure.go" -ErrorAction SilentlyContinue
    }
    
    default {
        Write-Host "Invalid choice. Please run the script again." -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "================================================" -ForegroundColor Cyan
