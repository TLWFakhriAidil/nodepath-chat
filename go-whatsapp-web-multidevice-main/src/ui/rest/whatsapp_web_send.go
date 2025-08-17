package rest

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
	"strings"
	"encoding/base64"
	"context"
	"io"
	"net/http"
	"go.mau.fi/whatsmeow"
	"os"
	"path/filepath"
	"github.com/sirupsen/logrus"
)

// SendWhatsAppWebMessage handles sending messages from WhatsApp Web view
func (handler *App) SendWhatsAppWebMessage(c *fiber.Ctx) error {
	deviceId := c.Params("id")
	
	// Message request structure
	type MessageRequest struct {
		ChatID   string `json:"chatId"`
		Message  string `json:"message"`
		ImageURL string `json:"imageUrl,omitempty"`
		ImageB64 string `json:"imageB64,omitempty"`
	}
	
	var request MessageRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request format",
		})
	}
	
	// Validate request
	if request.ChatID == "" || (request.Message == "" && request.ImageURL == "" && request.ImageB64 == "") {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST", 
			Message: "ChatID and either message or image is required",
		})
	}
	
	// Get session from cookie
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "No session token found",
		})
	}
	
	// Verify session and device ownership
	userRepo := repository.GetUserRepository()
	session, err := userRepo.GetSession(sessionToken)
	if err != nil || session == nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Invalid session",
		})
	}
	
	// Get user from database
	user, err := userRepo.GetUserByID(session.UserID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
	}
	
	// Get user devices to check if this device belongs to user
	devices, err := userRepo.GetUserDevices(user.ID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get devices",
		})
	}
	
	// Check if device belongs to user and is online
	deviceBelongsToUser := false
	isOnline := false
	for _, device := range devices {
		if device.ID == deviceId {
			deviceBelongsToUser = true
			isOnline = device.Status == "online"
			break
		}
	}
	
	if !deviceBelongsToUser {
		return c.Status(403).JSON(utils.ResponseData{
			Status:  403,
			Code:    "FORBIDDEN",
			Message: "Device does not belong to user",
		})
	}
	
	if !isOnline {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "DEVICE_OFFLINE",
			Message: "Device is offline",
		})
	}
	
	// Get WhatsApp client
	cm := whatsapp.GetClientManager()
	client, err := cm.GetClient(deviceId)
	if err != nil || client == nil || !client.IsConnected() {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "DEVICE_OFFLINE",
			Message: "Device is not connected",
		})
	}
	
	// Parse JID for recipient
	recipientJID, err := types.ParseJID(request.ChatID)
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "INVALID_JID",
			Message: fmt.Sprintf("Invalid chat ID format: %v", err),
		})
	}
	
	// Send message based on type
	if request.ImageURL != "" || request.ImageB64 != "" {
		// Send image message
		var imageData []byte
		
		// Handle base64 image
		if request.ImageB64 != "" {
			// Extract base64 data (remove data:image/jpeg;base64, prefix if present)
			b64Data := request.ImageB64
			if strings.Contains(b64Data, ",") {
				parts := strings.Split(b64Data, ",")
				if len(parts) > 1 {
					b64Data = parts[1]
				}
			}
			
			// Decode base64
			imageData, err = base64.StdEncoding.DecodeString(b64Data)
			if err != nil {
				return c.Status(400).JSON(utils.ResponseData{
					Status:  400,
					Code:    "INVALID_IMAGE",
					Message: "Failed to decode base64 image",
				})
			}
		} else if request.ImageURL != "" {
			// Download image from URL
			resp, err := http.Get(request.ImageURL)
			if err != nil {
				return c.Status(400).JSON(utils.ResponseData{
					Status:  400,
					Code:    "INVALID_IMAGE",
					Message: "Failed to download image from URL",
				})
			}
			defer resp.Body.Close()
			
			imageData, err = io.ReadAll(resp.Body)
			if err != nil {
				return c.Status(400).JSON(utils.ResponseData{
					Status:  400,
					Code:    "INVALID_IMAGE",
					Message: "Failed to read image data",
				})
			}
		}
		
		// Upload image to WhatsApp
		uploadResp, err := client.Upload(context.Background(), imageData, whatsmeow.MediaImage)
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "UPLOAD_FAILED",
				Message: fmt.Sprintf("Failed to upload image: %v", err),
			})
		}
		
		// Create image message
		imageMsg := &waE2E.ImageMessage{
			Caption:       proto.String(request.Message),
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			MediaKey:      uploadResp.MediaKey,
			Mimetype:      proto.String(http.DetectContentType(imageData)),
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileSHA256:    uploadResp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(imageData))),
		}
		
		// Send message
		resp, err := client.SendMessage(context.Background(), recipientJID, &waE2E.Message{
			ImageMessage: imageMsg,
		})
		
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "SEND_FAILED",
				Message: fmt.Sprintf("Failed to send image: %v", err),
			})
		}
		
		// Store in messages table with media URL
		// Save the image to disk first
		filename := fmt.Sprintf("%s.jpg", resp.ID)
		imagePath := filepath.Join(config.PathStorages, filename)
		err = os.WriteFile(imagePath, imageData, 0644)
		if err != nil {
			logrus.Errorf("Failed to save sent image: %v", err)
		}
		
		mediaURL := "/media/" + filename
		go whatsapp.StoreWhatsAppMessageWithMedia(deviceId, request.ChatID, resp.ID, client.Store.ID.String(), request.Message, "image", mediaURL)
		
		// Notify WebSocket
		go whatsapp.NotifyMessageUpdate(deviceId, request.ChatID, request.Message)
		
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Image sent successfully",
			Results: map[string]interface{}{
				"messageId": resp.ID,
				"status":    "sent",
				"imageUrl":  mediaURL,
			},
		})
		
	} else {
		// Send text message
		resp, err := client.SendMessage(c.UserContext(), recipientJID, &waE2E.Message{
			Conversation: proto.String(request.Message),
		})
		
		if err != nil {
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "SEND_FAILED",
				Message: fmt.Sprintf("Failed to send message: %v", err),
			})
		}
		
		// Store in messages table
		go whatsapp.StoreWhatsAppMessage(deviceId, request.ChatID, resp.ID, client.Store.ID.String(), request.Message, "text")
		
		// Notify WebSocket
		go whatsapp.NotifyMessageUpdate(deviceId, request.ChatID, request.Message)
		
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Message sent successfully",
			Results: map[string]interface{}{
				"messageId": resp.ID,
				"status":    "sent",
			},
		})
	}
}
