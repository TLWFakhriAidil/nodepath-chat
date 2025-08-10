package broadcast

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp/multidevice"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/antipattern"
	platform "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/external"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// WhatsAppMessageSender handles message sending with self-healing capabilities
type WhatsAppMessageSender struct {
	greetingProcessor  *antipattern.GreetingProcessor
	platformSender     *platform.PlatformSender
}

// NewWhatsAppMessageSender creates a new message sender
func NewWhatsAppMessageSender() *WhatsAppMessageSender {
	return &WhatsAppMessageSender{
		greetingProcessor: antipattern.NewGreetingProcessor(),
		platformSender:    platform.NewPlatformSender(),
	}
}

// SendMessage sends a message via WhatsApp with self-healing capabilities
// NOTE: Anti-spam is handled by BroadcastWorker, not here
func (w *WhatsAppMessageSender) SendMessage(deviceID string, msg *broadcast.BroadcastMessage) error {
	// Check if this is a platform device (Wablas/Whacenter)
	userRepo := repository.GetUserRepository()
	device, err := userRepo.GetDeviceByID(deviceID)
	if err != nil {
		logrus.Errorf("Failed to get device %s: %v", deviceID, err)
		return fmt.Errorf("device not found: %v", err)
	}
	
	// Only process line breaks, NO anti-spam here
	processedContent := w.processLineBreaks(msg.Content)
	msg.Message = processedContent
	msg.Content = processedContent
	
	// Debug log to see the processed content
	logrus.Debugf("Processed message content for %s: %s", msg.RecipientPhone, strings.ReplaceAll(processedContent, "\n", "\\n"))
	
	// Check if it's a platform device
	if device.Platform != "" {
		// logrus.Infof("Sending via platform %s for device %s", device.Platform, device.DeviceName)
		return w.platformSender.SendMessage(
			device.Platform,
			device.JID,  // JID contains the instance/token for platform devices
			msg.RecipientPhone,
			msg.RecipientName,
			msg.Message,
			msg.ImageURL,
			deviceID,
		)
	}
	
	// Normal WhatsApp Web sending
	return w.sendViaWhatsApp(deviceID, msg)
}

// processLineBreaks only processes line breaks, no anti-spam
func (w *WhatsAppMessageSender) processLineBreaks(content string) string {
	if content == "" {
		return content
	}
	
	// Debug log original content
	logrus.Debugf("Original content: %s", strings.ReplaceAll(content, "\n", "\\n"))
	
	// Fix line breaks for WhatsApp - ensure we have actual newline characters
	// Replace various line break representations with actual newlines
	content = strings.ReplaceAll(content, "\\n", "\n")
	content = strings.ReplaceAll(content, "%0A", "\n")
	content = strings.ReplaceAll(content, "%0a", "\n")
	content = strings.ReplaceAll(content, "<br>", "\n")
	content = strings.ReplaceAll(content, "<br/>", "\n")
	content = strings.ReplaceAll(content, "<br />", "\n")
	content = strings.ReplaceAll(content, "\r\n", "\n") // Windows line breaks
	content = strings.ReplaceAll(content, "\r", "\n")   // Old Mac line breaks
	
	// Debug log processed content
	logrus.Debugf("Processed content with line breaks: %s", strings.ReplaceAll(content, "\n", "\\n"))
	
	return content
}

// processMessageContent is deprecated - kept for backward compatibility
func (w *WhatsAppMessageSender) processMessageContent(msg *broadcast.BroadcastMessage, deviceID string) string {
	// Get the original content
	content := msg.Content
	if content == "" {
		content = msg.Message
	}
	
	// Only process line breaks
	return w.processLineBreaks(content)
}

// sendViaWhatsApp sends message via normal WhatsApp with self-healing client refresh
func (w *WhatsAppMessageSender) sendViaWhatsApp(deviceID string, msg *broadcast.BroadcastMessage) error {
	// 🔄 SELF-HEALING: Use DeviceManager for automatic refresh
	dm := multidevice.GetDeviceManager()
	waClient, err := dm.GetOrRefreshClient(deviceID)
	if err != nil {
		return fmt.Errorf("failed to get/refresh client for device %s: %v", deviceID, err)
	}
	
	// Double-check client health before sending
	if !dm.IsClientHealthy(waClient) {
		return fmt.Errorf("device %s client is not healthy after refresh", deviceID)
	}
	
	// No more keepalive or manual reconnection - client is guaranteed healthy
	logrus.Debugf("📤 Sending message via healthy client for device %s", deviceID)
	
	if !waClient.IsLoggedIn() {
		return fmt.Errorf("device %s is not logged in", deviceID)
	}
	
	// Parse recipient JID
	recipientJID, err := types.ParseJID(msg.RecipientPhone + "@s.whatsapp.net")
	if err != nil {
		// Try without suffix
		recipientJID, err = types.ParseJID(msg.RecipientPhone)
		if err != nil {
			return fmt.Errorf("invalid recipient phone: %v", err)
		}
	}
	
	// Validate recipient
	if !strings.ContainsRune(recipientJID.User, '@') {
		info, err := waClient.IsOnWhatsApp([]string{recipientJID.User})
		if err != nil {
			return fmt.Errorf("failed to check WhatsApp: %v", err)
		}
		if len(info) == 0 || !info[0].IsIn {
			return fmt.Errorf("recipient %s is not on WhatsApp", msg.RecipientPhone)
		}
	}
	
	// Send based on message type
	if msg.Type == "image" && msg.ImageURL != "" {
		return w.sendImageMessage(waClient, recipientJID, msg)
	} else {
		return w.sendTextMessage(waClient, recipientJID, msg)
	}
}

// sendTextMessage sends a text message
func (w *WhatsAppMessageSender) sendTextMessage(waClient *whatsmeow.Client, recipient types.JID, msg *broadcast.BroadcastMessage) error {
	// Use ExtendedTextMessage for better formatting support
	message := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(msg.Message),
		},
	}
	
	// Send message
	resp, err := waClient.SendMessage(context.Background(), recipient, message)
	if err != nil {
		return fmt.Errorf("failed to send text message: %v", err)
	}
	
	logrus.Infof("Text message sent to %s (ID: %s)", recipient.String(), resp.ID)
	return nil
}

// sendImageMessage sends an image message
func (w *WhatsAppMessageSender) sendImageMessage(waClient *whatsmeow.Client, recipient types.JID, msg *broadcast.BroadcastMessage) error {
	var imageData []byte
	var err error
	
	// Check if it's a data URL or regular URL
	if strings.HasPrefix(msg.ImageURL, "data:") {
		// Handle data URL (base64 encoded image)
		parts := strings.SplitN(msg.ImageURL, ",", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid data URL format")
		}
		
		// Decode base64 data
		imageData, err = base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return fmt.Errorf("failed to decode base64 image: %v", err)
		}
	} else {
		// Handle regular URL
		resp, err := http.Get(msg.ImageURL)
		if err != nil {
			return fmt.Errorf("failed to download image: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to download image: %s", resp.Status)
		}
		
		imageData, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read image data: %v", err)
		}
	}
	
	// Upload to WhatsApp servers
	uploaded, err := waClient.Upload(context.Background(), imageData, whatsmeow.MediaImage)
	if err != nil {
		return fmt.Errorf("failed to upload image: %v", err)
	}
	
	// Create image message with processed caption
	caption := msg.Message
	message := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Caption:       proto.String(caption),
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(imageData))),
			Mimetype:      proto.String("image/jpeg"),
		},
	}
	
	// Send message
	resp, err := waClient.SendMessage(context.Background(), recipient, message)
	if err != nil {
		return fmt.Errorf("failed to send image message: %v", err)
	}
	
	logrus.Infof("Image message sent to %s (ID: %s)", recipient.String(), resp.ID)
	return nil
}
