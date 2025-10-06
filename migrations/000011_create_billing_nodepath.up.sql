-- Create subscriptions_nodepath table for user subscriptions
CREATE TABLE IF NOT EXISTS subscriptions_nodepath (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    plan_name VARCHAR(255) NOT NULL DEFAULT 'Test Plan',
    plan_price DECIMAL(10,2) NOT NULL DEFAULT 1.00,
    plan_period VARCHAR(50) NOT NULL DEFAULT 'monthly',
    status ENUM('active', 'cancelled', 'suspended', 'pending') NOT NULL DEFAULT 'active',
    next_billing_date DATE NOT NULL,
    features JSON DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_next_billing_date (next_billing_date),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create payments_nodepath table for payment transactions and Billplz integration
CREATE TABLE IF NOT EXISTS payments_nodepath (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    subscription_id VARCHAR(255) DEFAULT NULL,
    bill_id VARCHAR(255) DEFAULT NULL COMMENT 'Billplz bill ID',
    invoice_number VARCHAR(255) UNIQUE NOT NULL COMMENT 'Our internal invoice number',
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'MYR',
    description TEXT DEFAULT NULL,
    status ENUM('pending', 'paid', 'failed', 'cancelled') NOT NULL DEFAULT 'pending',
    payment_method VARCHAR(50) DEFAULT 'billplz',
    billplz_url TEXT DEFAULT NULL COMMENT 'Payment URL from Billplz',
    paid_at TIMESTAMP NULL DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_subscription_id (subscription_id),
    INDEX idx_bill_id (bill_id),
    INDEX idx_invoice_number (invoice_number),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    
    FOREIGN KEY (subscription_id) REFERENCES subscriptions_nodepath(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create billing_history_nodepath table for simplified billing display
CREATE TABLE IF NOT EXISTS billing_history_nodepath (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    payment_id VARCHAR(255) NOT NULL,
    invoice_number VARCHAR(255) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'MYR',
    description TEXT NOT NULL,
    status ENUM('pending', 'paid', 'failed', 'cancelled') NOT NULL,
    payment_date DATE DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_payment_id (payment_id),
    INDEX idx_payment_date (payment_date),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    
    FOREIGN KEY (payment_id) REFERENCES payments_nodepath(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert test subscription data
INSERT IGNORE INTO subscriptions_nodepath (
    id, user_id, plan_name, plan_price, plan_period, status, next_billing_date, features
) VALUES (
    'test_sub_001',
    '00000000-0000-0000-0000-000000000001',
    'Test Plan',
    1.00,
    'monthly',
    'active',
    DATE_ADD(CURDATE(), INTERVAL 1 MONTH),
    JSON_ARRAY('WhatsApp Bot Integration', 'Flow Builder Access', 'Basic Analytics', 'Standard Support')
);

-- Insert test billing history data
INSERT IGNORE INTO payments_nodepath (
    id, user_id, subscription_id, invoice_number, amount, currency, description, status, paid_at
) VALUES 
(
    'pay_test_001',
    '00000000-0000-0000-0000-000000000001',
    'test_sub_001',
    CONCAT('INV-', DATE_FORMAT(NOW(), '%Y%m%d'), '-0001'),
    1.00,
    'MYR',
    'Test Plan - Monthly subscription',
    'paid',
    NOW()
),
(
    'pay_test_002',
    '00000000-0000-0000-0000-000000000001',
    'test_sub_001',
    CONCAT('INV-', DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 1 MONTH), '%Y%m%d'), '-0002'),
    1.00,
    'MYR',
    'Test Plan - Monthly subscription',
    'paid',
    DATE_SUB(NOW(), INTERVAL 1 MONTH)
),
(
    'pay_test_003',
    '00000000-0000-0000-0000-000000000001',
    'test_sub_001',
    CONCAT('INV-', DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 2 MONTH), '%Y%m%d'), '-0003'),
    1.00,
    'MYR',
    'Test Plan - Monthly subscription',
    'paid',
    DATE_SUB(NOW(), INTERVAL 2 MONTH)
);

-- Create corresponding billing history records
INSERT IGNORE INTO billing_history_nodepath (
    id, user_id, payment_id, invoice_number, amount, currency, description, status, payment_date
) VALUES 
(
    'hist_test_001',
    '00000000-0000-0000-0000-000000000001',
    'pay_test_001',
    CONCAT('INV-', DATE_FORMAT(NOW(), '%Y%m%d'), '-0001'),
    1.00,
    'MYR',
    'Test Plan - Monthly subscription',
    'paid',
    CURDATE()
),
(
    'hist_test_002',
    '00000000-0000-0000-0000-000000000001',
    'pay_test_002',
    CONCAT('INV-', DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 1 MONTH), '%Y%m%d'), '-0002'),
    1.00,
    'MYR',
    'Test Plan - Monthly subscription',
    'paid',
    DATE_SUB(CURDATE(), INTERVAL 1 MONTH)
),
(
    'hist_test_003',
    '00000000-0000-0000-0000-000000000001',
    'pay_test_003',
    CONCAT('INV-', DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 2 MONTH), '%Y%m%d'), '-0003'),
    1.00,
    'MYR',
    'Test Plan - Monthly subscription',
    'paid',
    DATE_SUB(CURDATE(), INTERVAL 2 MONTH)
);