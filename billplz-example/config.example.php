<?php
// RENAME THIS FILE TO config.php AND ADD YOUR ACTUAL VALUES
// NEVER COMMIT config.php TO GIT

return [
    'billplz' => [
        'api_key' => 'YOUR_BILLPLZ_API_KEY_HERE',
        'collection_id' => 'YOUR_COLLECTION_ID_HERE',
        'sandbox' => false, // Set to true for testing
    ],
    
    'database' => [
        'host' => 'localhost',
        'username' => 'YOUR_DB_USERNAME',
        'password' => 'YOUR_DB_PASSWORD',
        'database' => 'YOUR_DB_NAME',
    ],
    
    'urls' => [
        'callback' => 'https://yourdomain.com/callback.php',
        'redirect' => 'https://yourdomain.com/thank_you.php',
    ]
];