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
// Check for both VITE_ prefixed variables (for local dev) and regular DB_// Function to parse MYSQL_URI into connection parameters
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
        echo "Using MYSQL_URI for database connection\n";
    } else {
        echo "Failed to parse MYSQL_URI, using fallback values\n";
        $host = '157.245.206.124';
        $dbname = 'admin_railway';
        $user = 'admin_aqil';
        $pass = 'admin_aqil';
        $port = '3306';
    }
} else {
    echo "MYSQL_URI not found, using fallback values\n";
    $host = '157.245.206.124';
    $dbname = 'admin_railway';
    $user = 'admin_aqil';
    $pass = 'admin_aqil';
    $port = '3306';
}

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