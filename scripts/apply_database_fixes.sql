-- Emergency Database Fix Script for NodePath Chat
-- This script adds missing columns and tables to fix API errors
-- Safe to run multiple times

-- ==================================================
-- STEP 1: Fix user_nodepath table (Profile fixes)
-- ==================================================

-- Add missing profile columns to user_nodepath table
ALTER TABLE user_nodepath 
ADD COLUMN IF NOT EXISTS gmail VARCHAR(255) DEFAULT NULL COMMENT 'User Gmail address',
ADD COLUMN IF NOT EXISTS phone VARCHAR(20) DEFAULT NULL COMMENT 'User phone number',
ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'active' COMMENT 'User account status (active/inactive)',
ADD COLUMN IF NOT EXISTS expired TIMESTAMP NULL DEFAULT NULL COMMENT 'Account expiration date';

-- Create indexes for new columns
CREATE INDEX IF NOT EXISTS idx_user_nodepath_gmail ON user_nodepath(gmail);
CREATE INDEX IF NOT EXISTS idx_user_nodepath_phone ON user_nodepath(phone);
CREATE INDEX IF NOT EXISTS idx_user_nodepath_status ON user_nodepath(status);

-- Update existing users to have active status
UPDATE user_nodepath SET status = 'active' WHERE status IS NULL OR status = '';

-- ==================================================  
-- STEP 2: Create billing tables (Billing fixes)
-- ==================================================

-- Create subscriptions_nodepath table
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

-- Create payments_nodepath table
CREATE TABLE IF NOT EXISTS payments_nodepath (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    subscription_id VARCHAR(255) DEFAULT NULL,
    bill_id VARCHAR(255) DEFAULT NULL COMMENT 'Billplz bill ID',
    invoice_number VARCHAR(255) UNIQUE NOT NULL COMMENT 'Internal invoice number',
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

-- Create billing_history_nodepath table
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

-- ==================================================
-- STEP 3: Insert test data
-- ==================================================

-- Insert test subscription for the first user (ID=1)
INSERT IGNORE INTO subscriptions_nodepath (
    id, user_id, plan_name, plan_price, plan_period, status, next_billing_date, features
) VALUES (
    'test_sub_001',
    '1',
    'Test Plan',
    1.00,
    'monthly',
    'active',
    DATE_ADD(CURDATE(), INTERVAL 1 MONTH),
    JSON_ARRAY('WhatsApp Bot Integration', 'Flow Builder Access', 'Basic Analytics', 'Standard Support')
);

-- Insert test payment record
INSERT IGNORE INTO payments_nodepath (
    id, user_id, subscription_id, invoice_number, amount, currency, description, status, payment_method
) VALUES (
    'test_pay_001',
    '1', 
    'test_sub_001',
    'INV-TEST-001',
    1.00,
    'MYR',
    'Test Plan - Monthly subscription',
    'paid',
    'billplz'
);

-- Insert test billing history
INSERT IGNORE INTO billing_history_nodepath (
    id, user_id, payment_id, invoice_number, amount, currency, description, status, payment_date
) VALUES (
    'test_hist_001',
    '1',
    'test_pay_001', 
    'INV-TEST-001',
    1.00,
    'MYR',
    'Test Plan - Monthly subscription',
    'paid',
    CURDATE()
);

-- ==================================================
-- STEP 4: Verification queries
-- ==================================================

-- Uncomment these lines to verify the migration worked:

-- Check user_nodepath table structure
-- DESCRIBE user_nodepath;

-- Check if billing tables were created
-- SHOW TABLES LIKE '%_nodepath';

-- Check test data was inserted
-- SELECT 'Test subscription:' as info; 
-- SELECT * FROM subscriptions_nodepath LIMIT 1;

-- SELECT 'Test payment:' as info;
-- SELECT * FROM payments_nodepath LIMIT 1;

-- SELECT 'Test billing history:' as info;
-- SELECT * FROM billing_history_nodepath LIMIT 1;

-- Final success message
SELECT 'Database migration completed successfully! Profile and billing APIs should now work.' as result;