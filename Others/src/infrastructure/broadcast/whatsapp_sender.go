package broadcast

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
	
	domainBroadcast "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

// DownloadMedia downloads media from URL or data URL
func DownloadMedia(url string) ([]byte, error) {
	// Check if it's a data URL
	if strings.HasPrefix(url, "data:") {
		// Parse data URL
		parts := strings.SplitN(url, ",", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid data URL format")
		}
		
		// Decode base64
		data, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %v", err)
		}
		
		return data, nil
	}
	
	// Regular HTTP/HTTPS URL
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}
	
	return io.ReadAll(resp.Body)
}

// GetRandomDelay returns a random delay between min and max seconds
func GetRandomDelay(minDelay, maxDelay int) time.Duration {
	if minDelay <= 0 {
		minDelay = 5
	}
	if maxDelay <= 0 || maxDelay < minDelay {
		maxDelay = minDelay + 10
	}
	
	delay := rand.Intn(maxDelay-minDelay) + minDelay
	return time.Duration(delay) * time.Second
}

// SendWhatsAppMessage sends a message using the WhatsApp client
func SendWhatsAppMessage(client *whatsmeow.Client, msg domainBroadcast.BroadcastMessage) error {
	// Sanitize phone number
	phone := msg.RecipientPhone
	whatsapp.SanitizePhone(&phone)
	
	recipient, err := types.ParseJID(phone + "@s.whatsapp.net")
	if err != nil {
		logrus.Errorf("Invalid phone number %s: %v", phone, err)
		return err
	}
	
	logrus.Infof("Sending %s message to %s", msg.Type, recipient.String())
	
	// Handle different message types
	switch msg.Type {
	case "text":
		if msg.MediaURL != "" {
			// Image with caption
			logrus.Infof("Sending image with caption to %s", recipient.String())
			
			imageBytes, err := DownloadMedia(msg.MediaURL)
			if err != nil {
				logrus.Errorf("Failed to download image: %v", err)
				return err
			}
			
			uploadResp, err := client.Upload(context.Background(), imageBytes, whatsmeow.MediaImage)
			if err != nil {
				logrus.Errorf("Failed to upload image: %v", err)
				return err
			}
			
			imageMsg := &waProto.ImageMessage{
				Caption:       proto.String(msg.Content),
				URL:           proto.String(uploadResp.URL),
				DirectPath:    proto.String(uploadResp.DirectPath),
				MediaKey:      uploadResp.MediaKey,
				Mimetype:      proto.String("image/jpeg"),
				FileEncSHA256: uploadResp.FileEncSHA256,
				FileSHA256:    uploadResp.FileSHA256,
				FileLength:    proto.Uint64(uint64(len(imageBytes))),
			}
			
			_, err = client.SendMessage(context.Background(), recipient, &waProto.Message{
				ImageMessage: imageMsg,
			})
			
			if err != nil {
				logrus.Errorf("Failed to send image message: %v", err)
				return err
			}
			
			logrus.Infof("Successfully sent image with caption to %s", recipient.String())
		} else {
			// Text only
			logrus.Infof("Sending text message to %s", recipient.String())
			
			_, err = client.SendMessage(context.Background(), recipient, &waProto.Message{
				Conversation: proto.String(msg.Content),
			})
			
			if err != nil {
				logrus.Errorf("Failed to send text message: %v", err)
				return err
			}
			
			logrus.Infof("Successfully sent text message to %s", recipient.String())
		}
		
	case "image":
		// Image only (no caption)
		logrus.Infof("Sending image to %s", recipient.String())
		
		imageBytes, err := DownloadMedia(msg.MediaURL)
		if err != nil {
			logrus.Errorf("Failed to download image: %v", err)
			return err
		}
		
		uploadResp, err := client.Upload(context.Background(), imageBytes, whatsmeow.MediaImage)
		if err != nil {
			logrus.Errorf("Failed to upload image: %v", err)
			return err
		}
		
		imageMsg := &waProto.ImageMessage{
			URL:           proto.String(uploadResp.URL),
			DirectPath:    proto.String(uploadResp.DirectPath),
			MediaKey:      uploadResp.MediaKey,
			Mimetype:      proto.String("image/jpeg"),
			FileEncSHA256: uploadResp.FileEncSHA256,
			FileSHA256:    uploadResp.FileSHA256,
			FileLength:    proto.Uint64(uint64(len(imageBytes))),
		}
		
		_, err = client.SendMessage(context.Background(), recipient, &waProto.Message{
			ImageMessage: imageMsg,
		})
		
		if err != nil {
			logrus.Errorf("Failed to send image: %v", err)
			return err
		}
		
		logrus.Infof("Successfully sent image to %s", recipient.String())
		
	default:
		logrus.Warnf("Unknown message type: %s", msg.Type)
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
	
	// Wait for the delay
	delay := GetRandomDelay(msg.MinDelay, msg.MaxDelay)
	logrus.Infof("Waiting %v before next message", delay)
	time.Sleep(delay)
	
	return nil
}
