	Balas           sql.NullString  `json:"balas" db:"balas"` // VARCHAR(255) for timestamp
	DataImage       string          `json:"data_image" db:"data_image"`
	ConvStage       string          `json:"conv_stage" db:"conv_stage"`
	Niche           string          `json:"niche" db:"niche"`
	BotBalas        *time.Time      `json:"bot_balas" db:"bot_balas"`