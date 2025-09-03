	IDDevice        string          `json:"id_device" db:"id_device"`
	ProspectNum     string          `json:"prospect_num" db:"prospect_num"`
	Stage           string          `json:"stage" db:"stage"`
	DateOrder       *time.Time      `json:"date_order" db:"date_order"`
	ConvLast        json.RawMessage `json:"conv_last" db:"conv_last"`
	ConvCurrent     sql.NullString  `json:"conv_current" db:"conv_current"`
	Jam             string          `json:"jam" db:"jam"`
	Intro           string          `json:"intro" db:"intro"`
	Human           int             `json:"human" db:"human"` // 0 = AI active, 1 = human takeover
	CatatanStaff    string          `json:"catatan_staff" db:"catatan_staff"`
	Balas           int             `json:"balas" db:"balas"` // Will be changed to string after migration
	DataImage       string          `json:"data_image" db:"data_image"`
	ConvStage       string          `json:"conv_stage" db:"conv_stage"`
	Niche           string          `json:"niche" db:"niche"`
	BotBalas        *time.Time      `json:"bot_balas" db:"bot_balas"`
	KeywordIklan    string          `json:"keywordiklan" db:"keywordiklan"`
	Marketer        string          `json:"marketer" db:"marketer"`
	UpdateToday     *time.Time      `json:"update_today" db:"update_today"`
	// Flow execution fields
	FlowReference   sql.NullString  `json:"flow_reference" db:"flow_reference"`   // Reference to chatbot flow being executed
	CurrentNode     sql.NullString  `json:"current_node" db:"current_node"`       // Current node in the flow execution
	Variables       json.RawMessage `json:"variables" db:"variables"`             // Flow execution variables (JSON)
	ExecutionStatus sql.NullString  `json:"execution_status" db:"execution_status"` // Flow execution status (active, completed, failed)
	ExecutionID     sql.NullString  `json:"execution_id" db:"execution_id"`       // Unique execution identifier
	// Flow tracking fields for user reply handling
	CurrentNodeID   sql.NullString  `json:"current_node_id" db:"current_node_id"`   // Current node ID in the chatbot flow
	WaitingForReply sql.NullInt32   `json:"waiting_for_reply" db:"waiting_for_reply"` // 1 = waiting for user reply, 0 = not waiting
	FlowID          sql.NullString  `json:"flow_id" db:"flow_id"`                 // ID of the current chatbot flow being executed
	LastNodeID      sql.NullString  `json:"last_node_id" db:"last_node_id"`       // Previous node ID for flow tracking
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`