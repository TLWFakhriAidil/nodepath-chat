-- Create team member device assignments table
CREATE TABLE IF NOT EXISTS team_member_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_member_id UUID NOT NULL REFERENCES team_members(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES user_devices(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by UUID REFERENCES users(id),
    UNIQUE(team_member_id, device_id)
);

-- Create index for faster lookups
CREATE INDEX idx_team_member_devices_member ON team_member_devices(team_member_id);
CREATE INDEX idx_team_member_devices_device ON team_member_devices(device_id);
