// Direct MySQL API without Supabase
const MYSQL_CONFIG = {
  host: '159.89.198.71',
  port: 3306,
  user: 'admin_aqil',
  password: 'admin_aqil',
  database: 'admin_railway'
}

// Direct MySQL API call using a custom backend endpoint
export const callDirectMySQLAPI = async (query: string, params: any[] = []) => {
  try {
    console.log('Direct MySQL Query:', query, 'Params:', params);
    
    // Since we can't connect directly from browser to MySQL, we'll use a custom backend
    // For now, we'll create a PHP backend endpoint that you can deploy
    const response = await fetch('https://your-backend-domain.com/mysql-api.php', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query,
        params,
        config: MYSQL_CONFIG
      })
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const result = await response.json();
    
    if (result.success) {
      console.log('Direct MySQL operation successful:', result);
      return result;
    } else {
      throw new Error(result.error || 'MySQL operation failed');
    }
  } catch (error) {
    console.error('Direct MySQL connection error:', error)
    throw error
  }
}

// Alternative: Use a direct MySQL connection library for Node.js
// This would require running a separate Node.js backend server
export const createBackendAPI = () => {
  return `
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

    if (empty($params)) {
        $stmt = $pdo->query($query);
        $result = $stmt->fetchAll();
    } else {
        $stmt = $pdo->prepare($query);
        $stmt->execute($params);
        $result = $stmt->fetchAll();
    }

    echo json_encode([
        'success' => true,
        'data' => $result,
        'affectedRows' => $stmt->rowCount()
    ]);

} catch (PDOException $e) {
    http_response_code(500);
    echo json_encode([
        'success' => false,
        'error' => $e->getMessage()
    ]);
}
?>
  `;
}