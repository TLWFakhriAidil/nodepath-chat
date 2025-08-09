<?php
// Enable error reporting for debugging
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Set CORS headers
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: GET, OPTIONS');
header('Access-Control-Allow-Headers: Content-Type, Accept');
header('Content-Type: application/json');

// Create a log function for debugging
function debug_log($message) {
  error_log($message);
}

debug_log('Request received: ' . $_SERVER['REQUEST_METHOD']);

// Handle preflight requests
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
  http_response_code(200);
  exit;
}

try {
    debug_log('Attempting to connect to MySQL');
    
    // MySQL connection parameters
    $host = '159.89.198.71';
    $port = 3306;
    $database = 'admin_railway';
    $user = 'admin_aqil';
    $password = 'admin_aqil';
    
    $dsn = "mysql:host={$host};port={$port};dbname={$database};charset=utf8mb4";
    
    debug_log('DSN: ' . $dsn);
    
    $pdo = new PDO($dsn, $user, $password, [
        PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
        PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
        PDO::ATTR_EMULATE_PREPARES => false,
    ]);
    
    debug_log('Connected to MySQL successfully');

    // Simple test query
    $stmt = $pdo->query('SELECT 1 as test');
    $result = $stmt->fetchAll();
    
    $response = [
        'success' => true,
        'message' => 'MySQL connection successful',
        'data' => $result,
        'server_info' => [
            'php_version' => phpversion(),
            'server_software' => $_SERVER['SERVER_SOFTWARE'] ?? 'Unknown',
            'request_time' => date('Y-m-d H:i:s'),
            'remote_addr' => $_SERVER['REMOTE_ADDR'] ?? 'Unknown'
        ]
    ];
    
    echo json_encode($response, JSON_PRETTY_PRINT);
    debug_log('Response sent successfully');

} catch (PDOException $e) {
    debug_log('PDO Exception: ' . $e->getMessage());
    debug_log('Error code: ' . $e->getCode());
    
    // Return detailed error information
    http_response_code(500);
    $errorResponse = [
        'success' => false, 
        'error' => $e->getMessage(),
        'errorCode' => $e->getCode(),
        'server_info' => [
            'php_version' => phpversion(),
            'server_software' => $_SERVER['SERVER_SOFTWARE'] ?? 'Unknown',
            'request_time' => date('Y-m-d H:i:s'),
            'remote_addr' => $_SERVER['REMOTE_ADDR'] ?? 'Unknown'
        ]
    ];
    
    echo json_encode($errorResponse, JSON_PRETTY_PRINT);
    debug_log('Error response sent');
}
?>