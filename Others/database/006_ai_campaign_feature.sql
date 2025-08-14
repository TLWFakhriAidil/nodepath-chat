-- AI Campaign Feature Migration
-- This migration adds support for AI-powered lead management and campaign distribution

-- 1. Create leads_ai table for AI-managed leads
CREATE TABLE IF NOT EXISTS leads_ai (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    device_id VARCHAR(255), -- Initially NULL, assigned during campaign
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    email VARCHAR(255),
    niche VARCHAR(255),
    source VARCHAR(255) DEFAULT 'ai_manual',
    status VARCHAR(50) DEFAULT 'pending', -- pending, assigned, sent, failed
    target_status VARCHAR(50) DEFAULT 'prospect', -- prospect/customer
    notes TEXT,
    assigned_at TIMESTAMP, -- When assigned to device
    sent_at TIMESTAMP, -- When message was sent
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_leads_ai_user_id ON leads_ai(user_id);
CREATE INDEX idx_leads_ai_device_id ON leads_ai(device_id);
CREATE INDEX idx_leads_ai_status ON leads_ai(status);
CREATE INDEX idx_leads_ai_niche ON leads_ai(niche);
CREATE INDEX idx_leads_ai_phone ON leads_ai(phone);

-- 2. Add AI columns to campaigns table
ALTER TABLE campaigns 
ADD COLUMN IF NOT EXISTS ai VARCHAR(10),
ADD COLUMN IF NOT EXISTS "limit" INTEGER DEFAULT 0;

-- 3. Create ai_campaign_progress table for tracking device usage
CREATE TABLE IF NOT EXISTS ai_campaign_progress (
    id SERIAL PRIMARY KEY,
    campaign_id INTEGER NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    leads_sent INTEGER DEFAULT 0,
    leads_failed INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'active', -- active, limit_reached, failed
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE,
    UNIQUE(campaign_id, device_id)
);

-- Create indexes
CREATE INDEX idx_ai_campaign_progress_campaign_id ON ai_campaign_progress(campaign_id);
CREATE INDEX idx_ai_campaign_progress_device_id ON ai_campaign_progress(device_id);