package send

type MessageRequest struct {
	DeviceID       string  `json:"device_id" form:"device_id"`
	Phone          string  `json:"phone" form:"phone"`
	Message        string  `json:"message" form:"message"`
	IsForwarded    bool    `json:"is_forwarded" form:"is_forwarded"`
	ReplyMessageID *string `json:"reply_message_id" form:"reply_message_id"`
}
