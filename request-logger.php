<?php
// Request logger script to capture all incoming requests
header('Content-Type: application/json');

// Enable error reporting
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Create a log file with timestamp
$logFile = './php-requests-' . date('Y-m-d') . '.log';

// Log function
function log_request($message) {
    global $logFile;
    $timestamp = date('Y-m-d H:i:s');
    file_put_contents($logFile, "[$timestamp] $message\n", FILE_APPEND);
}

// Log basic request information
log_request('=== New Request ===');
log_request('Method: ' . ($_SERVER['REQUEST_METHOD'] ?? 'UNKNOWN'));
log_request('URI: ' . ($_SERVER['REQUEST_URI'] ?? 'UNKNOWN'));

// Log headers
log_request('=== Headers ===');
foreach (getallheaders() as $name => $value) {
    log_request("$name: $value");
}

// Log raw input
$rawInput = file_get_contents('php://input');
log_request('=== Raw Input ===');
log_request($rawInput);

// Log server variables
log_request('=== Server Variables ===');
foreach ($_SERVER as $key => $value) {
    if (!is_array($value)) {
        log_request("$key: $value");
    }
}

// Log environment variables
log_request('=== Environment Variables ===');
foreach ($_ENV as $key => $value) {
    log_request("$key: $value");
}

// Prepare response
$response = [
    'success' => true,
    'message' => 'Request logged successfully',
    'timestamp' => date('Y-m-d H:i:s'),
    'log_file' => $logFile
];

// Output response
echo json_encode($response, JSON_PRETTY_PRINT);