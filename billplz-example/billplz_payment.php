<?php

ini_set('display_errors', 1);
ini_set('display_startup_errors', 1);
error_reporting(E_ALL);

// Load configuration
$config = require_once 'config.php';

// Get the customer details and amount from the form submission
$customer_email = isset($_POST['billing_email']) ? $_POST['billing_email'] : 'customer@example.com';
$customer_name = isset($_POST['billing_first_name']) ? $_POST['billing_first_name'] : 'Customer Name';
$billing_phone = isset($_POST['billing_phone']) ? $_POST['billing_phone'] : 'Unknown';
$billing_address = isset($_POST['billing_address_1']) ? $_POST['billing_address_1'] : 'Unknown';
$billing_city = isset($_POST['billing_city']) ? $_POST['billing_city'] : 'Unknown';
$billing_state = isset($_POST['billing_state']) ? $_POST['billing_state'] : 'Unknown';
$billing_postcode = isset($_POST['billing_postcode']) ? $_POST['billing_postcode'] : 'Unknown';
$amount = isset($_POST['amount']) ? intval($_POST['amount']) : 3000; // Billplz requires amount in cents
$product_name = isset($_POST['product_name']) ? $_POST['product_name'] : 'Unknown Product';
$payment_method = isset($_POST['payment_method']) ? $_POST['payment_method'] : 'billplz';
$description = $product_name;

// Log form data for debugging
file_put_contents('form_data_log.txt', print_r($_POST, true), FILE_APPEND);

// Database connection using config
$mysqli = new mysqli(
    $config['database']['host'],
    $config['database']['username'],
    $config['database']['password'],
    $config['database']['database']
);

if ($mysqli->connect_error) {
    file_put_contents('error_log.txt', "Connection failed: " . $mysqli->connect_error . PHP_EOL, FILE_APPEND);
    die("Connection failed: " . $mysqli->connect_error);
}

$status = $payment_method === 'cod' ? 'Pending' : 'Processing';
$amount_in_rm = $amount / 100; // Convert amount from cents to RM
$bill_id_placeholder = ''; // Placeholder for bill_id

// Insert order into database
$stmt = $mysqli->prepare("INSERT INTO orders (customer_email, customer_name, billing_phone, billing_address, billing_city, billing_state, billing_postcode, amount, collection_id, status, bill_id, product, method) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)");
if ($stmt === false) {
    file_put_contents('error_log.txt', "Prepare failed: " . $mysqli->error . PHP_EOL, FILE_APPEND);
    die("Prepare failed: " . $mysqli->error);
}

if (!$stmt->bind_param("sssssssisssss", $customer_email, $customer_name, $billing_phone, $billing_address, $billing_city, $billing_state, $billing_postcode, $amount_in_rm, $config['billplz']['collection_id'], $status, $bill_id_placeholder, $product_name, $payment_method)) {
    file_put_contents('error_log.txt', "Binding parameters failed: " . $stmt->error . PHP_EOL, FILE_APPEND);
    die("Binding parameters failed: " . $stmt->error);
}

if (!$stmt->execute()) {
    file_put_contents('error_log.txt', "Execute failed: " . $stmt->error . PHP_EOL, FILE_APPEND);
    die("Execute failed: " . $stmt->error);
}

// Get the last inserted order ID
$order_id = $stmt->insert_id;

$stmt->close();

// Remove the corresponding entry from abandoned_leads if phone number matches
$remove_lead_stmt = $mysqli->prepare("DELETE FROM abandoned_leads WHERE billing_phone = ?");
if ($remove_lead_stmt === false) {
    file_put_contents('error_log.txt', "Prepare failed: " . $mysqli->error . PHP_EOL, FILE_APPEND);
    die("Prepare failed: " . $mysqli->error);
}

if (!$remove_lead_stmt->bind_param("s", $billing_phone)) {
    file_put_contents('error_log.txt', "Binding parameters failed: " . $remove_lead_stmt->error . PHP_EOL, FILE_APPEND);
    die("Binding parameters failed: " . $remove_lead_stmt->error);
}

if (!$remove_lead_stmt->execute()) {
    file_put_contents('error_log.txt', "Execute failed: " . $remove_lead_stmt->error . PHP_EOL, FILE_APPEND);
    die("Execute failed: " . $remove_lead_stmt->error);
}

$remove_lead_stmt->close();
$mysqli->close();

// If the payment method is COD, just redirect to thank you page
if ($payment_method === 'cod') {
    header('Location: ' . $config['urls']['redirect'] . $order_id);
    exit();
}

// Proceed with Billplz payment if the payment method is online banking
$data = array(
    'collection_id' => $config['billplz']['collection_id'],
    'email' => $customer_email,
    'name' => $customer_name,
    'amount' => $amount,
    'description' => $description,
    'callback_url' => $config['urls']['callback'],
    'redirect_url' => $config['urls']['redirect'] . $order_id,
    'reference_1' => (string)$order_id,
    'reference_1_label' => 'order_id'
);

// Determine Billplz API URL based on sandbox setting
$billplz_url = $config['billplz']['sandbox'] 
    ? 'https://www.billplz-sandbox.com/api/v3/bills'
    : 'https://www.billplz.com/api/v3/bills';

$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, $billplz_url);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
curl_setopt($ch, CURLOPT_POST, 1);
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($data));
curl_setopt($ch, CURLOPT_HTTPHEADER, array(
    'Content-Type: application/json',
    'Authorization: Basic ' . base64_encode($config['billplz']['api_key'] . ':')
));

$response = curl_exec($ch);
if ($response === false) {
    file_put_contents('error_log.txt', "Curl error: " . curl_error($ch) . PHP_EOL, FILE_APPEND);
    die("Curl error: " . curl_error($ch));
}
curl_close($ch);

$result = json_decode($response, true);

if (isset($result['id']) && isset($result['url'])) {
    $bill_id = $result['id'];
    $billplz_url = $result['url'];

    // Update order with bill_id
    $mysqli = new mysqli(
        $config['database']['host'],
        $config['database']['username'],
        $config['database']['password'],
        $config['database']['database']
    );
    
    if ($mysqli->connect_error) {
        file_put_contents('error_log.txt', "Connection failed: " . $mysqli->connect_error . PHP_EOL, FILE_APPEND);
        die("Connection failed: " . $mysqli->connect_error);
    }

    $stmt = $mysqli->prepare("UPDATE orders SET bill_id = ?, url = ? WHERE id = ?");
    if ($stmt === false) {
        file_put_contents('error_log.txt', "Prepare failed: " . $mysqli->error . PHP_EOL, FILE_APPEND);
        die("Prepare failed: " . $mysqli->error);
    }

    if (!$stmt->bind_param("ssi", $bill_id, $billplz_url, $order_id)) {
        file_put_contents('error_log.txt', "Binding parameters failed: " . $stmt->error . PHP_EOL, FILE_APPEND);
        die("Binding parameters failed: " . $stmt->error);
    }

    if (!$stmt->execute()) {
        file_put_contents('error_log.txt', "Execute failed: " . $stmt->error . PHP_EOL, FILE_APPEND);
        die("Execute failed: " . $stmt->error);
    }

    $stmt->close();
    $mysqli->close();

    header('Location: ' . $billplz_url);
    exit();
} else {
    file_put_contents('error_log.txt', "Billplz error: " . (isset($result['error']['message']) ? $result['error']['message'] : 'Unknown error') . PHP_EOL, FILE_APPEND);
    echo 'Error: ' . (isset($result['error']['message']) ? $result['error']['message'] : 'Unknown error');
}