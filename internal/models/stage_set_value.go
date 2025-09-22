package models

import (
	"database/sql"
	"time"
)

// StageSetValue represents the stageSetValue_nodepath table
type StageSetValue struct {
	StageSetValueID int            `json:"stageSetValue_id" db:"stageSetValue_id"`
	IDDevice        string         `json:"id_device" db:"id_device"`
	Stage           int            `json:"stage" db:"stage"`
	TypeInputData   string         `json:"type_inputData" db:"type_inputData"`
	ColumnsData     string         `json:"columnsData" db:"columnsData"`
	InputHardCode   sql.NullString `json:"inputHardCode" db:"inputHardCode"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" db:"updated_at"`
}
