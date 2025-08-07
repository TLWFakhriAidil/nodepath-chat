<?php
header("Access-Control-Allow-Origin: *");
header("Access-Control-Allow-Methods: POST, OPTIONS");
header("Access-Control-Allow-Headers: Content-Type");
header("Content-Type: application/json");

if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    exit(0);
}

if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    http_response_code(405);
    echo json_encode(['success' => false, 'error' => 'Method not allowed']);
    exit;
}

$input = json_decode(file_get_contents('php://input'), true);

if (!$input) {
    http_response_code(400);
    echo json_encode(['success' => false, 'error' => 'Invalid JSON']);
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
    $dsn = "mysql:host={$config['host']};port={$config['port']};dbname={$config['database']};charset=utf8mb4";
    $pdo = new PDO($dsn, $config['user'], $config['password'], [
        PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
        PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC,
    ]);

    // Handle multiple statements (separated by semicolon)
    $statements = array_filter(array_map('trim', explode(';', $query)));
    $result = [];
    $affectedRows = 0;
    
    foreach ($statements as $i => $statement) {
        if (empty($statement)) continue;
        
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
    }

    echo json_encode([
        'success' => true,
        'data' => $result,
        'affectedRows' => $affectedRows
    ]);

} catch (PDOException $e) {
    http_response_code(500);
    echo json_encode([
        'success' => false,
        'error' => $e->getMessage()
    ]);
}
?>
