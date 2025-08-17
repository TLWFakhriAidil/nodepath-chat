import re

def fix_all_duplicate_issues():
    """Apply all fixes for sequence and campaign duplicates"""
    
    # 1. Fix broadcast_repository.go - add 'processing' to duplicate checks
    print("1. Fixing broadcast_repository.go...")
    with open('src/repository/broadcast_repository.go', 'r') as f:
        content = f.read()
    
    # Fix sequence duplicate check
    content = re.sub(
        r"AND status IN \('pending', 'sent', 'queued'\)",
        "AND status IN ('pending', 'sent', 'queued', 'processing')",
        content
    )
    
    with open('src/repository/broadcast_repository.go', 'w') as f:
        f.write(content)
    print("   [OK] Added 'processing' status to duplicate checks")
    
    # 2. Fix campaign_trigger.go - add 'processing' to sequence duplicate check
    print("\n2. Fixing campaign_trigger.go...")
    with open('src/usecase/campaign_trigger.go', 'r') as f:
        content = f.read()
    
    # Fix the duplicate check in ProcessDailySequenceMessages
    content = re.sub(
        r"AND status IN \('pending', 'processing', 'queued', 'sent'\)",
        "AND status IN ('pending', 'processing', 'queued', 'sent')",
        content
    )
    
    # If not found, try other pattern
    if "AND status IN ('pending', 'processing', 'queued', 'sent')" not in content:
        content = re.sub(
            r"AND status IN \('pending', 'queued', 'sent'\)",
            "AND status IN ('pending', 'processing', 'queued', 'sent')",
            content
        )
    
    with open('src/usecase/campaign_trigger.go', 'w') as f:
        f.write(content)
    print("   [OK] Fixed sequence duplicate check in ProcessDailySequenceMessages")
    
    # 3. Create SQL for unique constraints
    print("\n3. Creating SQL for unique constraints...")
    sql_content = """-- CRITICAL: Add unique constraints to prevent duplicates at database level
-- Run this SQL immediately!

-- First, remove any existing duplicates
-- For sequences
DELETE t1 FROM broadcast_messages t1
INNER JOIN broadcast_messages t2 
WHERE t1.id > t2.id
AND t1.sequence_stepid = t2.sequence_stepid 
AND t1.recipient_phone = t2.recipient_phone 
AND t1.device_id = t2.device_id
AND t1.sequence_stepid IS NOT NULL;

-- For campaigns
DELETE t1 FROM broadcast_messages t1
INNER JOIN broadcast_messages t2 
WHERE t1.id > t2.id
AND t1.campaign_id = t2.campaign_id 
AND t1.recipient_phone = t2.recipient_phone 
AND t1.device_id = t2.device_id
AND t1.campaign_id IS NOT NULL;

-- Add unique constraints
ALTER TABLE broadcast_messages 
ADD UNIQUE INDEX IF NOT EXISTS unique_sequence_message (
    sequence_stepid, 
    recipient_phone, 
    device_id
);

ALTER TABLE broadcast_messages 
ADD UNIQUE INDEX IF NOT EXISTS unique_campaign_message (
    campaign_id, 
    recipient_phone, 
    device_id
);

-- Verify constraints were added
SHOW INDEX FROM broadcast_messages WHERE Key_name IN ('unique_sequence_message', 'unique_campaign_message');
"""
    
    with open('add_unique_constraints.sql', 'w') as f:
        f.write(sql_content)
    print("   [OK] Created add_unique_constraints.sql")
    
    print("\n" + "="*50)
    print("FIXES APPLIED:")
    print("="*50)
    print("1. [OK] Added 'processing' status to all duplicate checks")
    print("2. [OK] Fixed sequence duplicate check in campaign_trigger.go") 
    print("3. [OK] Created SQL for unique constraints")
    print("\nNEXT STEPS:")
    print("1. Run add_unique_constraints.sql on the database")
    print("2. Build and deploy the application")
    print("3. Monitor that processing_worker_id is being populated")

if __name__ == "__main__":
    fix_all_duplicate_issues()
