<?php
// Enable error reporting for debugging
ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Set CORS headers
header('Access-Control-Allow-Origin: *');
header('Access-Control-Allow-Methods: POST, OPTIONS');
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

if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    http_response_code(405);
    echo json_encode(['success' => false, 'error' => 'Method not allowed']);
    exit;
}

// Get raw input and log it for debugging
$rawInput = file_get_contents('php://input');
file_put_contents('php://stderr', "Raw input received: " . $rawInput . "\n");

// Check if input is empty
if (empty($rawInput)) {
    http_response_code(400);
    echo json_encode(['success' => false, 'error' => 'Empty input received']);
    exit;
}

// Parse JSON input with error handling
try {
    $input = json_decode($rawInput, true, 512, JSON_THROW_ON_ERROR);
    file_put_contents('php://stderr', "Successfully parsed JSON input\n");
} catch (Exception $e) {
    // Log detailed error information
    $errorMsg = 'JSON parse error: ' . $e->getMessage();
    $inputLength = strlen($rawInput);
    $inputPreview = substr($rawInput, 0, 100) . ($inputLength > 100 ? '...' : '');
    
    file_put_contents('php://stderr', $errorMsg . "\n");
    file_put_contents('php://stderr', "Input length: " . $inputLength . "\n");
    file_put_contents('php://stderr', "Input preview: " . $inputPreview . "\n");
    
    http_response_code(400);
    echo json_encode(['success' => false, 'error' => $errorMsg]);
    exit;
}

// Check if input is null or not an array
if (!$input || !is_array($input)) {
    file_put_contents('php://stderr', "Invalid JSON structure\n");
    http_response_code(400);
    echo json_encode(['success' => false, 'error' => 'Invalid JSON structure']);
    exit;
}

$query = $input['query'] ?? '';
$params = $input['params'] ?? [];
$config = $input['config'] ?? [];

if (empty($query)) {
    http_response_code(400);
    echo json_encode(['success' => false, 'error' => 'Query is required']);
    exit;
}

try {
    debug_log('Attempting to connect to MySQL: ' . $config['host']);
    
    $port = isset($config['port']) ? $config['port'] : 3306;
    $dsn = "mysql:host={$config['host']};port={$port};dbname={$config['database']};charset=utf8mb4";
    
    debug_log('DSN: ' . $dsn);
    
    $pdo = new PDO($dsn, $config['user'], $config['password'], [
        PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
        PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
        PDO::ATTR_EMULATE_PREPARES => false,
    ]);
    
    debug_log('Connected to MySQL successfully');

    // Handle multiple statements (separated by semicolon)
    $statements = array_filter(array_map('trim', explode(';', $query)));
    $result = [];
    $affectedRows = 0;
    
    debug_log('Executing query with ' . count($statements) . ' statements');
    debug_log('With parameters: ' . json_encode($params));
    
    foreach ($statements as $i => $statement) {
        if (empty($statement)) continue;
        
        debug_log('Executing statement ' . ($i+1) . ': ' . $statement);
        
        if ($i === count($statements) - 1 && !empty($params)) {
            // Last statement with parameters
            $stmt = $pdo->prepare($statement);
            $stmt->execute($params);
        } else {
            // Other statements without parameters
            $stmt = $pdo->prepare($statement);
            $stmt->execute();
        }
        
        $result = $stmt->fetchAll();
        $affectedRows += $stmt->rowCount();
        debug_log('Statement executed successfully, fetched ' . count($result) . ' rows');
    }

    $response = [
        'success' => true,
        'data' => $result,
        'affectedRows' => $affectedRows
    ];
    echo json_encode($response);
    debug_log('Response sent successfully');

} catch (PDOException $e) {
    debug_log('PDO Exception: ' . $e->getMessage());
    debug_log('Error code: ' . $e->getCode());
    
    // Return detailed error information
    http_response_code(500);
    $errorResponse = [
        'success' => false, 
        'error' => $e->getMessage(),
        'errorCode' => $e->getCode()
    ];
    
    echo json_encode($errorResponse);
    debug_log('Error response sent');
}
?>
