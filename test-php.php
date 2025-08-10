<?php
// Simple test script to verify PHP is working
header('Content-Type: application/json');

// Return basic information about PHP
$info = [
    'success' => true,
    'php_version' => PHP_VERSION,
    'server' => $_SERVER,
    'extensions' => get_loaded_extensions(),
    'pdo_drivers' => PDO::getAvailableDrivers(),
    'timestamp' => date('Y-m-d H:i:s')
];

echo json_encode($info, JSON_PRETTY_PRINT);