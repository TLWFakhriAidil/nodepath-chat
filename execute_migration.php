<?php
/**
 * Manual Migration Executor for Railway Production Database
 * This script adds all missing columns to ai_whatsapp_nodepath table
 */

header('Content-Type: application/json');
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: GET, POST, OPTIONS');
header('Access-Control-Allow-Headers: Content-Type');

// Database connection parameters
$host = '159.89.198.71';
$port = 3306;
$database = 'admin_railway';
$username = 'admin_aqil';
$password = 'admin_aqil';

try {
    // Create PDO connection
    $dsn = "mysql:host=$host;port=$port;dbname=$database;charset=utf8mb4";
    $pdo = new PDO($dsn, $username, $password, [
        PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
        PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
        PDO::ATTR_TIMEOUT => 30
    ]);
    
    echo json_encode(['status' => 'success', 'message' => 'Connected to database successfully']);
    echo "\n";
    
    // Check current table structure
    echo json_encode(['status' => 'info', 'message' => 'Checking current table structure...']);
    echo "\n";
    
    $stmt = $pdo->query("DESCRIBE ai_whatsapp_nodepath");
    $currentColumns = $stmt->fetchAll();
    
    $existingColumns = array_column($currentColumns, 'Field');
    echo json_encode(['status' => 'info', 'message' => 'Current columns: ' . implode(', ', $existingColumns)]);
    echo "\n";
    
    // Define columns to add
    $columnsToAdd = [
        'jam' => "ADD COLUMN jam VARCHAR(255) DEFAULT NULL COMMENT 'Jam field for AI WhatsApp conversations'",
        'intro' => "ADD COLUMN intro VARCHAR(255) DEFAULT NULL COMMENT 'Introduction field'",
        'date_order' => "ADD COLUMN date_order DATETIME DEFAULT NULL COMMENT 'Order date field'",
        'balas' => "ADD COLUMN balas VARCHAR(255) DEFAULT NULL COMMENT 'Reply field'",
        'data_image' => "ADD COLUMN data_image TEXT DEFAULT NULL COMMENT 'Image data field'",
        'conv_stage' => "ADD COLUMN conv_stage VARCHAR(100) DEFAULT NULL COMMENT 'Conversation stage field'",
        'keywordiklan' => "ADD COLUMN keywordiklan VARCHAR(255) DEFAULT NULL COMMENT 'Advertisement keyword field'",
        'marketer' => "ADD COLUMN marketer VARCHAR(255) DEFAULT NULL COMMENT 'Marketer field'",
        'update_today' => "ADD COLUMN update_today TINYINT(1) DEFAULT 0 COMMENT 'Update today flag'"
    ];
    
    // Add missing columns
    $addedColumns = [];
    $skippedColumns = [];
    
    foreach ($columnsToAdd as $columnName => $alterStatement) {
        if (!in_array($columnName, $existingColumns)) {
            try {
                $pdo->exec("ALTER TABLE ai_whatsapp_nodepath $alterStatement");
                $addedColumns[] = $columnName;
                echo json_encode(['status' => 'success', 'message' => "Added column: $columnName"]);
                echo "\n";
            } catch (PDOException $e) {
                echo json_encode(['status' => 'error', 'message' => "Failed to add column $columnName: " . $e->getMessage()]);
                echo "\n";
            }
        } else {
            $skippedColumns[] = $columnName;
            echo json_encode(['status' => 'info', 'message' => "Column $columnName already exists, skipping"]);
            echo "\n";
        }
    }
    
    // Fix data types for existing columns
    echo json_encode(['status' => 'info', 'message' => 'Fixing data types for existing columns...']);
    echo "\n";
    
    try {
        $pdo->exec("ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN id_prospect INT DEFAULT NULL COMMENT 'Prospect ID as integer'");
        echo json_encode(['status' => 'success', 'message' => 'Fixed id_prospect data type to INT']);
        echo "\n";
    } catch (PDOException $e) {
        echo json_encode(['status' => 'warning', 'message' => 'Could not modify id_prospect: ' . $e->getMessage()]);
        echo "\n";
    }
    
    try {
        $pdo->exec("ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN bot_balas TIMESTAMP NULL DEFAULT NULL COMMENT 'Bot reply timestamp'");
        echo json_encode(['status' => 'success', 'message' => 'Fixed bot_balas data type to TIMESTAMP']);
        echo "\n";
    } catch (PDOException $e) {
        echo json_encode(['status' => 'warning', 'message' => 'Could not modify bot_balas: ' . $e->getMessage()]);
        echo "\n";
    }
    
    // Verify final table structure
    echo json_encode(['status' => 'info', 'message' => 'Verifying final table structure...']);
    echo "\n";
    
    $stmt = $pdo->query("DESCRIBE ai_whatsapp_nodepath");
    $finalColumns = $stmt->fetchAll();
    
    $finalColumnNames = array_column($finalColumns, 'Field');
    echo json_encode(['status' => 'info', 'message' => 'Final columns: ' . implode(', ', $finalColumnNames)]);
    echo "\n";
    
    // Summary
    echo json_encode([
        'status' => 'success',
        'message' => 'Migration completed successfully!',
        'summary' => [
            'added_columns' => $addedColumns,
            'skipped_columns' => $skippedColumns,
            'total_columns' => count($finalColumnNames)
        ]
    ]);
    echo "\n";
    
} catch (PDOException $e) {
    echo json_encode([
        'status' => 'error',
        'message' => 'Database connection failed: ' . $e->getMessage()
    ]);
    echo "\n";
} catch (Exception $e) {
    echo json_encode([
        'status' => 'error',
        'message' => 'Unexpected error: ' . $e->getMessage()
    ]);
    echo "\n";
}
?>