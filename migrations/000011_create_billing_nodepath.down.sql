-- Drop billing tables in reverse order due to foreign key constraints
DROP TABLE IF EXISTS billing_history_nodepath;
DROP TABLE IF EXISTS payments_nodepath;  
DROP TABLE IF EXISTS subscriptions_nodepath;