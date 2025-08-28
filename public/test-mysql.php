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
    // Get database connection from environment variables
$mysqlURI = getenv('MYSQL_URI');
if (!$mysqlURI) {
    die('MYSQL_URI environment variable not set');
}

// Parse MYSQL_URI
if (!preg_match('/mysql:\/\/([^:]+):([^@]+)@([^:]+):(\d+)\/(.+)/', $mysqlURI, $matches)) {
    die('Invalid MYSQL_URI format');
}

$user = $matches[1];
$password = $matches[2];
$host = $matches[3];
$port = (int)$matches[4];
$database = $matches[5];
    
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
    
    // Ensure proper JSON encoding with error handling
    $jsonResponse = json_encode($response, JSON_PRETTY_PRINT | JSON_PARTIAL_OUTPUT_ON_ERROR);
    if ($jsonResponse === false) {
        // If JSON encoding fails, send a simple success response
        debug_log('JSON encoding failed: ' . json_last_error_msg());
        echo '{"success":true,"message":"MySQL connection successful","data":[{"test":1}]}'; 
    } else {
        echo $jsonResponse;
    }
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
    
    // Ensure proper JSON encoding with error handling
    $jsonResponse = json_encode($errorResponse, JSON_PRETTY_PRINT | JSON_PARTIAL_OUTPUT_ON_ERROR);
    if ($jsonResponse === false) {
        // If JSON encoding fails, send a simple error response
        debug_log('JSON encoding failed: ' . json_last_error_msg());
        echo '{"success":false,"error":"Internal server error: ' . addslashes(json_last_error_msg()) . '"}'; 
    } else {
        echo $jsonResponse;
    }
    debug_log('Error response sent');
}
?>