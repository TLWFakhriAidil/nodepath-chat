<?php
// Database connection test script
header('Content-Type: application/json');

// Enable error reporting
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Log function
function debug_log($message) {
    error_log($message);
}

debug_log('Starting database connection test');

// Get database connection parameters from environment variables
// Check for both VITE_ prefixed variables (for local dev) and regular DB_ variables (for Railway)
$host = getenv('VITE_DB_HOST') ?: getenv('DB_HOST') ?: '159.89.198.71';
$dbname = getenv('VITE_DB_NAME') ?: getenv('DB_NAME') ?: 'admin_railway';
$user = getenv('VITE_DB_USER') ?: getenv('DB_USER') ?: 'admin_aqil';
$pass = getenv('VITE_DB_PASSWORD') ?: getenv('DB_PASSWORD') ?: 'admin_aqil';
$port = getenv('VITE_DB_PORT') ?: getenv('DB_PORT') ?: '3306';

debug_log("Connection parameters: host=$host, dbname=$dbname, user=$user");

// Try to connect to the database
try {
    // Construct DSN
    $dsn = "mysql:host={$host};port={$port};dbname={$dbname}";
    debug_log("DSN: $dsn");
    
    // Set PDO options
    $options = [
        PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
        PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
        PDO::ATTR_EMULATE_PREPARES => false,
    ];
    
    // Create PDO instance
    debug_log("Creating PDO instance");
    $pdo = new PDO($dsn, $user, $pass, $options);
    debug_log("PDO instance created successfully");
    
    // Test query
    $stmt = $pdo->query("SELECT 1 AS test");
    $result = $stmt->fetch();
    
    // Prepare response
    $response = [
        'success' => true,
        'message' => 'Database connection successful',
        'test_result' => $result,
        'connection_info' => [
            'host' => $host,
            'database' => $dbname,
            'user' => $user,
            'driver' => $pdo->getAttribute(PDO::ATTR_DRIVER_NAME),
            'server_version' => $pdo->getAttribute(PDO::ATTR_SERVER_VERSION),
            'client_version' => $pdo->getAttribute(PDO::ATTR_CLIENT_VERSION)
        ]
    ];
} catch (PDOException $e) {
    debug_log("PDO Exception: " . $e->getMessage());
    
    // Prepare error response
    $response = [
        'success' => false,
        'message' => 'Database connection failed',
        'error' => $e->getMessage(),
        'error_code' => $e->getCode(),
        'connection_info' => [
            'host' => $host,
            'database' => $dbname,
            'user' => $user
        ]
    ];
}

// Output response
echo json_encode($response, JSON_PRETTY_PRINT);