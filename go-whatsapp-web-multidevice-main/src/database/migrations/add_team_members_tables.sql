-- Create team members table
CREATE TABLE IF NOT EXISTS team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,  -- matches device name
    password VARCHAR(255) NOT NULL,         -- plain text (as requested)
    created_by UUID,                        -- leader's user_id
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT true
);

-- Create team sessions table
CREATE TABLE IF NOT EXISTS team_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_member_id UUID REFERENCES team_members(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_team_members_username ON team_members(username);
CREATE INDEX IF NOT EXISTS idx_team_members_created_by ON team_members(created_by);
CREATE INDEX IF NOT EXISTS idx_team_sessions_token ON team_sessions(token);
CREATE INDEX IF NOT EXISTS idx_team_sessions_expires_at ON team_sessions(expires_at);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_team_members_updated_at BEFORE UPDATE
    ON team_members FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
