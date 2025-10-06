-- ================================
-- URGENT DATABASE FIX FOR NODEPATH CHAT
-- Run this in phpMyAdmin SQL tab
-- ================================

-- Step 1: Add missing columns to user_nodepath
ALTER TABLE `user_nodepath` 
ADD COLUMN `gmail` VARCHAR(255) NULL DEFAULT NULL COMMENT 'User Gmail address' AFTER `full_name`,
ADD COLUMN `phone` VARCHAR(20) NULL DEFAULT NULL COMMENT 'User phone number' AFTER `gmail`;

-- Update existing users to have active status if NULL
UPDATE `user_nodepath` SET `status` = 'active' WHERE `status` IS NULL OR `status` = '';

-- Step 2: Create missing billing tables
CREATE TABLE `subscriptions_nodepath` (
  `id` varchar(255) NOT NULL,
  `user_id` varchar(255) NOT NULL,
  `plan_name` varchar(255) NOT NULL DEFAULT 'Test Plan',
  `plan_price` decimal(10,2) NOT NULL DEFAULT 1.00,
  `plan_period` varchar(50) NOT NULL DEFAULT 'monthly',
  `status` enum('active','cancelled','suspended','pending') NOT NULL DEFAULT 'active',
  `next_billing_date` date NOT NULL,
  `features` json DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_next_billing_date` (`next_billing_date`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `payments_nodepath` (
  `id` varchar(255) NOT NULL,
  `user_id` varchar(255) NOT NULL,
  `subscription_id` varchar(255) DEFAULT NULL,
  `bill_id` varchar(255) DEFAULT NULL COMMENT 'Billplz bill ID',
  `invoice_number` varchar(255) NOT NULL COMMENT 'Internal invoice number',
  `amount` decimal(10,2) NOT NULL,
  `currency` varchar(10) NOT NULL DEFAULT 'MYR',
  `description` text DEFAULT NULL,
  `status` enum('pending','paid','failed','cancelled') NOT NULL DEFAULT 'pending',
  `payment_method` varchar(50) DEFAULT 'billplz',
  `billplz_url` text DEFAULT NULL COMMENT 'Payment URL from Billplz',
  `paid_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `invoice_number` (`invoice_number`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_subscription_id` (`subscription_id`),
  KEY `idx_bill_id` (`bill_id`),
  KEY `idx_invoice_number` (`invoice_number`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `payments_nodepath_ibfk_1` FOREIGN KEY (`subscription_id`) REFERENCES `subscriptions_nodepath` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `billing_history_nodepath` (
  `id` varchar(255) NOT NULL,
  `user_id` varchar(255) NOT NULL,
  `payment_id` varchar(255) NOT NULL,
  `invoice_number` varchar(255) NOT NULL,
  `amount` decimal(10,2) NOT NULL,
  `currency` varchar(10) NOT NULL DEFAULT 'MYR',
  `description` text NOT NULL,
  `status` enum('pending','paid','failed','cancelled') NOT NULL,
  `payment_date` date DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_payment_id` (`payment_id`),
  KEY `idx_payment_date` (`payment_date`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `billing_history_nodepath_ibfk_1` FOREIGN KEY (`payment_id`) REFERENCES `payments_nodepath` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Step 3: Insert test data for user ID 1 (assuming you have a user with ID 1)
INSERT IGNORE INTO `subscriptions_nodepath` 
(`id`, `user_id`, `plan_name`, `plan_price`, `plan_period`, `status`, `next_billing_date`, `features`) 
VALUES 
('test_sub_001', '1', 'Test Plan', 1.00, 'monthly', 'active', DATE_ADD(CURDATE(), INTERVAL 1 MONTH), 
JSON_ARRAY('WhatsApp Bot Integration', 'Flow Builder Access', 'Basic Analytics', 'Standard Support'));

INSERT IGNORE INTO `payments_nodepath` 
(`id`, `user_id`, `subscription_id`, `invoice_number`, `amount`, `currency`, `description`, `status`, `payment_method`) 
VALUES 
('test_pay_001', '1', 'test_sub_001', 'INV-TEST-001', 1.00, 'MYR', 'Test Plan - Monthly subscription', 'paid', 'billplz');

INSERT IGNORE INTO `billing_history_nodepath` 
(`id`, `user_id`, `payment_id`, `invoice_number`, `amount`, `currency`, `description`, `status`, `payment_date`) 
VALUES 
('test_hist_001', '1', 'test_pay_001', 'INV-TEST-001', 1.00, 'MYR', 'Test Plan - Monthly subscription', 'paid', CURDATE());

-- Step 4: Verification queries (run these to check success)
SELECT 'User table columns:' as info;
DESCRIBE user_nodepath;

SELECT 'Billing tables created:' as info;
SHOW TABLES LIKE '%_nodepath';

SELECT 'Test subscription data:' as info;
SELECT * FROM subscriptions_nodepath LIMIT 1;

SELECT 'MIGRATION COMPLETED SUCCESSFULLY!' as result;