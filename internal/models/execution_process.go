package models

import "time"

// ExecutionProcess represents a conversation execution lock to prevent duplicate parallel processing
type ExecutionProcess struct {
	IDChatInput int       `db:"id_chatInput" json:"id_chatInput"`
	IDDevice    string    `db:"id_device" json:"id_device"`
	IDProspect  string    `db:"id_prospect" json:"id_prospect"`
	Times       time.Time `db:"times" json:"times"`
}

// TableName returns the table name for ExecutionProcess
func (ExecutionProcess) TableName() string {
	return "execution_process_nodepath"
}
