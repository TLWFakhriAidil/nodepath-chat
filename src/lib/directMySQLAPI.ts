// Direct MySQL API without Supabase - Using Railway environment variables
// Configuration is fetched from backend API to avoid exposing credentials in frontend
let MYSQL_CONFIG: any = null;

// Function to fetch database configuration from backend
const fetchDatabaseConfig = async () => {
  if (MYSQL_CONFIG) return MYSQL_CONFIG;
  
  try {
    const response = await fetch('/api/config/database');
    if (response.ok) {
      MYSQL_CONFIG = await response.json();
      return MYSQL_CONFIG;
    }
  } catch (error) {
    console.warn('Failed to fetch database config from backend, using fallback');
  }
  
  // Fallback configuration (should not contain real credentials)
  MYSQL_CONFIG = {
    host: 'localhost',
    port: 3306,
    user: 'root',
    password: '',
    database: 'admin_railway'
  };
  
  return MYSQL_CONFIG;
}

// Direct MySQL API call using a custom backend endpoint
export const callDirectMySQLAPI = async (query: string, params: any[] = []) => {
  try {
    console.log('Direct MySQL Query:', query, 'Params:', params);
    
    // Fetch database configuration from backend
    const config = await fetchDatabaseConfig();
    
    // Prepare the request payload
    const payload = {
      query,
      params,
      config
    };
    
    console.log('Sending payload:', JSON.stringify(payload));
    
    // Since we can't connect directly from browser to MySQL, we'll use a custom backend
    const response = await fetch('/mysql-api.php', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      },
      body: JSON.stringify(payload)
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    // Get response as text first to handle potential JSON parsing issues
    const responseText = await response.text();
    
    // Check if response is empty
    if (!responseText) {
      throw new Error('Empty response from server');
    }
    
    // Try to parse the response as JSON
    let result;
    try {
      result = JSON.parse(responseText);
    } catch (parseError) {
      console.error('Failed to parse JSON response:', responseText);
      throw new Error(`JSON parse error: ${parseError.message}`);
    }
    
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