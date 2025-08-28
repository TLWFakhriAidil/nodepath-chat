<?php
// Debug script to help identify issues with the PHP API
header('Content-Type: application/json');

// Enable error reporting
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Log function
function debug_log($message) {
    error_log($message);
}

// Get request information
$method = $_SERVER['REQUEST_METHOD'] ?? 'UNKNOWN';
debug_log("Request method: {$method}");

// Get raw input
$rawInput = file_get_contents('php://input');
debug_log("Raw input: {$rawInput}");

// Get server variables
$serverVars = [];
foreach ($_SERVER as $key => $value) {
    if (!is_array($value)) {
        $serverVars[$key] = $value;
    }
}

// Check MySQL connection
$mysqlStatus = 'Not tested';
try {
    // Function to parse MYSQL_URI into connection parameters
    function parseMySQLURI($mysqlURI) {
        if (empty($mysqlURI) || !str_starts_with($mysqlURI, 'mysql://')) {
            return null;
        }
        
        // Remove mysql:// prefix
        $url = substr($mysqlURI, 8);
        
        // Split user:password@host:port/database
        $parts = explode('@', $url);
        if (count($parts) !== 2) {
            return null;
        }
        
        $userPass = $parts[0];
        $hostPortDB = $parts[1];
        
        // Split user:password
        $userParts = explode(':', $userPass);
        if (count($userParts) !== 2) {
            return null;
        }
        
        $user = $userParts[0];
        $password = $userParts[1];
        
        // Split host:port/database
        $hostParts = explode('/', $hostPortDB);
        if (count($hostParts) !== 2) {
            return null;
        }
        
        $database = $hostParts[1];
        $hostPort = $hostParts[0];
        
        $hostPortParts = explode(':', $hostPort);
        if (count($hostPortParts) !== 2) {
            return null;
        }
        
        $host = $hostPortParts[0];
        $port = $hostPortParts[1];
        
        return [
            'host' => $host,
            'port' => $port,
            'user' => $user,
            'password' => $password,
            'database' => $database
        ];
    }

    // Database configuration using MYSQL_URI exclusively
    $mysqlURI = getenv('MYSQL_URI');
    if ($mysqlURI) {
        $config = parseMySQLURI($mysqlURI);
        if ($config) {
            $host = $config['host'];
            $dbname = $config['database'];
            $user = $config['user'];
            $pass = $config['password'];
            $port = $config['port'];
        } else {
            $host = '157.245.206.124';
            $dbname = 'admin_railway';
            $user = 'admin_aqil';
            $pass = 'admin_aqil';
            $port = '3306';
        }
    } else {
        $host = '157.245.206.124';
        $dbname = 'admin_railway';
        $user = 'admin_aqil';
        $pass = 'admin_aqil';
        $port = '3306';
    }
    
    $dsn = "mysql:host={$host};dbname={$dbname}";
    $options = [
        PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
        PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
        PDO::ATTR_EMULATE_PREPARES => false,
    ];
    
    $pdo = new PDO($dsn, $user, $pass, $options);
    $mysqlStatus = 'Connected successfully';
} catch (PDOException $e) {
    $mysqlStatus = 'Connection failed: ' . $e->getMessage();
}

// Check PHP extensions
$extensions = get_loaded_extensions();
$hasPDO = in_array('pdo', $extensions);
$hasPDOMySQL = in_array('pdo_mysql', $extensions);
$hasJSON = in_array('json', $extensions);

// Check file permissions
$apiFile = __DIR__ . '/mysql-api.php';
$apiFileExists = file_exists($apiFile);
$apiFilePermissions = $apiFileExists ? substr(sprintf('%o', fileperms($apiFile)), -4) : 'N/A';

// Prepare response
$response = [
    'success' => true,
    'timestamp' => date('Y-m-d H:i:s'),
    'php_version' => PHP_VERSION,
    'request_method' => $method,
    'raw_input_length' => strlen($rawInput),
    'raw_input' => $rawInput,
    'server_variables' => $serverVars,
    'mysql_status' => $mysqlStatus,
    'extensions' => [
        'all' => $extensions,
        'pdo' => $hasPDO,
        'pdo_mysql' => $hasPDOMySQL,
        'json' => $hasJSON
    ],
    'file_info' => [
        'api_file_exists' => $apiFileExists,
        'api_file_permissions' => $apiFilePermissions,
        'current_dir' => __DIR__,
        'files_in_dir' => scandir(__DIR__)
    ],
    'environment' => $_ENV
];

// Output response
echo json_encode($response, JSON_PRETTY_PRINT);