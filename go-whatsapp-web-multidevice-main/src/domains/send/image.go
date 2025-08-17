package send

import "mime/multipart"

type ImageRequest struct {
	DeviceID    string                `json:"device_id" form:"device_id"`
	Phone       string                `json:"phone" form:"phone"`
	Caption     string                `json:"caption" form:"caption"`
	Image       *multipart.FileHeader `json:"image" form:"image"`
	ImageURL    string                `json:"image_url" form:"image_url"`
	ImageB64    string                `json:"image_b64" form:"image_b64"`
	ImageBytes  []byte                `json:"-"`
	ViewOnce    bool                  `json:"view_once" form:"view_once"`
	Compress    bool                  `json:"compress"`
	IsForwarded bool                  `json:"is_forwarded" form:"is_forwarded"`
}
