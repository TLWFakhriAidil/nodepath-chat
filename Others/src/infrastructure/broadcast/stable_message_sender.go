package broadcast

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp/stability"
	platform "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/external"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// StableMessageSender handles WhatsApp message sending with ULTRA stability
type StableMessageSender struct {
	ultraStable     *stability.UltraStableConnection
	platformSender  *platform.PlatformSender
	userRepo        *repository.UserRepository
}

// NewStableMessageSender creates a new stable message sender
func NewStableMessageSender() *StableMessageSender {
	return &StableMessageSender{
		ultraStable:    stability.GetUltraStableConnection(),
		platformSender: platform.NewPlatformSender(),
		userRepo:      repository.GetUserRepository(),
	}
}

// SendMessage sends a message with MAXIMUM stability - no disconnections allowed
func (s *StableMessageSender) SendMessage(deviceID string, msg *broadcast.BroadcastMessage) error {
	// Get device details
	device, err := s.userRepo.GetDeviceByID(deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device: %v", err)
	}
	
	// Platform devices - always stable
	if device.Platform != "" {
		logrus.Infof("Sending via platform %s (always stable)", device.Platform)
		return s.platformSender.SendMessage(
			device.Platform,
			device.JID,
			msg.RecipientPhone,
			msg.RecipientName,
			msg.Message,
			msg.ImageURL,
			deviceID,
		)
	}
	
	// Get ultra-stable client
	waClient, err := s.ultraStable.GetStableClient(deviceID)
	if err != nil {
		// If not registered for ultra-stable, register it now
		logrus.Warnf("Device %s not in ultra-stable mode, registering now", deviceID)
		
		// Try to get from normal client manager first
		cm := whatsapp.GetClientManager()
		normalClient, err := cm.GetClient(deviceID)
		if err != nil {
			return fmt.Errorf("device not available: %v", err)
		}
		
		// Register for ultra-stable
		s.ultraStable.RegisterClient(deviceID, normalClient)
		
		// Now get the stable client
		waClient, err = s.ultraStable.GetStableClient(deviceID)
		if err != nil {
			return fmt.Errorf("failed to get stable client: %v", err)
		}
	}
	
	// At this point, client MUST be connected (ultra-stable ensures this)
	if !waClient.IsConnected() {
		logrus.Errorf("CRITICAL: Ultra-stable client %s not connected - this should never happen!", deviceID)
		// Force one more connection attempt
		waClient.Connect()
		time.Sleep(1 * time.Second)
	}
	
	// Parse recipient
	recipientJID, err := types.ParseJID(msg.RecipientPhone + "@s.whatsapp.net")
	if err != nil {
		recipientJID, err = types.ParseJID(msg.RecipientPhone)
		if err != nil {
			return fmt.Errorf("invalid recipient: %v", err)
		}
	}
	
	// Send the message - NO DELAYS, MAXIMUM SPEED
	if msg.Type == "image" && msg.ImageURL != "" {
		return s.sendImageMaxSpeed(waClient, recipientJID, msg)
	} else {
		return s.sendTextMaxSpeed(waClient, recipientJID, msg)
	}
}

// sendTextMaxSpeed sends text at maximum speed
func (s *StableMessageSender) sendTextMaxSpeed(waClient *whatsmeow.Client, recipient types.JID, msg *broadcast.BroadcastMessage) error {
	message := &waE2E.Message{
		Conversation: proto.String(msg.Message),
	}
	
	// Send with no validation or delays
	resp, err := waClient.SendMessage(context.Background(), recipient, message)
	if err != nil {
		logrus.Errorf("Send failed for %s: %v - device might be banned", recipient, err)
		return err
	}
	
	logrus.Debugf("Message sent at MAX SPEED to %s (ID: %s)", recipient, resp.ID)
	return nil
}

// sendImageMaxSpeed sends image at maximum speed
func (s *StableMessageSender) sendImageMaxSpeed(waClient *whatsmeow.Client, recipient types.JID, msg *broadcast.BroadcastMessage) error {
	// Quick image handling - no validation
	var imageData []byte
	
	if strings.HasPrefix(msg.ImageURL, "data:") {
		parts := strings.SplitN(msg.ImageURL, ",", 2)
		if len(parts) == 2 {
			imageData, _ = base64.StdEncoding.DecodeString(parts[1])
		}
	} else {
		resp, err := http.Get(msg.ImageURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		imageData, _ = io.ReadAll(resp.Body)
	}
	
	// Upload fast
	uploaded, err := waClient.Upload(context.Background(), imageData, whatsmeow.MediaImage)
	if err != nil {
		return err
	}
	
	// Send immediately
	message := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Caption:       proto.String(msg.Message),
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(imageData))),
			Mimetype:      proto.String("image/jpeg"),
		},
	}
	
	resp, err := waClient.SendMessage(context.Background(), recipient, message)
	if err != nil {
		return err
	}
	
	logrus.Debugf("Image sent at MAX SPEED to %s (ID: %s)", recipient, resp.ID)
	return nil
}
