-- Add gmail and phone columns to user_nodepath table
ALTER TABLE user_nodepath 
ADD COLUMN gmail VARCHAR(255) DEFAULT NULL COMMENT 'User Gmail address',
ADD COLUMN phone VARCHAR(20) DEFAULT NULL COMMENT 'User phone number';

-- Add indexes for the new columns
CREATE INDEX idx_user_nodepath_gmail ON user_nodepath(gmail);
CREATE INDEX idx_user_nodepath_phone ON user_nodepath(phone);