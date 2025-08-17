package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
	// Read the UltraOptimizedBroadcastProcessor file
	filePath := "src/usecase/ultra_optimized_broadcast_processor.go"
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	fileContent := string(content)

	// Replace the entire processMessages function to use GetPendingMessagesAndLock
	oldProcessMessages := `func (p *UltraOptimizedBroadcastProcessor) processMessages() {
	db := database.GetDB()
	
	// Get pending messages grouped by broadcast
	rows, err := db.Query(` + "`" + `
		SELECT bm.id, bm.user_id, bm.device_id, bm.campaign_id, bm.sequence_id,
			bm.recipient_phone, bm.recipient_name, bm.content AS message, bm.media_url AS image_url,
			COALESCE(c.min_delay_seconds, 5) AS min_delay,
			COALESCE(c.max_delay_seconds, 15) AS max_delay,
			COALESCE(d.status, 'unknown') AS device_status,
			COALESCE(d.platform, '') AS platform
		FROM broadcast_messages bm
		LEFT JOIN campaigns c ON bm.campaign_id = c.id
		LEFT JOIN user_devices d ON bm.device_id = d.id
		WHERE bm.status = 'pending'
		AND bm.scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
		ORDER BY bm.campaign_id, bm.sequence_id, bm.created_at
		LIMIT 1000
	` + "`" + `)
	
	if err != nil {
		logrus.Errorf("Failed to get pending messages: %v", err)
		return
	}
	defer rows.Close()
	
	messageCount := 0
	campaignPools := make(map[int]bool)
	sequencePools := make(map[string]bool)
	
	for rows.Next() {
		var msg domainBroadcast.BroadcastMessage
		var campaignID *int
		var sequenceID *string
		var minDelay, maxDelay int
		var deviceStatus sql.NullString
		var devicePlatform string
		var imageURL sql.NullString // Use sql.NullString for nullable fields
		var recipientName sql.NullString // Add recipient name
		
		err := rows.Scan(
			&msg.ID, &msg.UserID, &msg.DeviceID, &campaignID, &sequenceID,
			&msg.RecipientPhone, &recipientName, &msg.Message, &imageURL, // Include recipient name
			&minDelay, &maxDelay, &deviceStatus, &devicePlatform,
		)
		
		if err != nil {
			logrus.Errorf("Failed to scan message: %v", err)
			continue
		}
		
		// Set recipient name
		if recipientName.Valid {
			msg.RecipientName = recipientName.String
		} else {
			msg.RecipientName = ""
		}
		
		// Convert NullString to string
		if imageURL.Valid {
			msg.ImageURL = imageURL.String
		} else {
			msg.ImageURL = ""
		}
		
		// Check device status - platform devices are always considered online
		if devicePlatform == "" && deviceStatus.String != "connected" && deviceStatus.String != "online" {
			// Skip this WhatsApp Web device - mark messages as skipped
			db.Exec(` + "`" + `UPDATE broadcast_messages SET status = 'skipped', error_message = 'Device offline' 
					 WHERE device_id = ? AND status = 'pending'` + "`" + `, msg.DeviceID)
			continue
		}
		
		// Set broadcast references
		msg.CampaignID = campaignID
		msg.SequenceID = sequenceID
		msg.MinDelay = minDelay
		msg.MaxDelay = maxDelay
		msg.Type = "text"
		if msg.ImageURL != "" {
			msg.Type = "image"
		}
		
		// Create pool if needed
		if campaignID != nil && !campaignPools[*campaignID] {
			_, err := p.manager.StartBroadcastPool("campaign", fmt.Sprintf("%d", *campaignID))
			if err != nil {
				logrus.Errorf("Failed to start campaign pool: %v", err)
				continue
			}
			campaignPools[*campaignID] = true
			
			// Update campaign status to processing
			db.Exec(` + "`" + `UPDATE campaigns SET status = 'processing', 
					 updated_at = NOW() 
					 WHERE id = ?` + "`" + `, *campaignID)
		}
		
		if sequenceID != nil && !sequencePools[*sequenceID] {
			_, err := p.manager.StartBroadcastPool("sequence", *sequenceID)
			if err != nil {
				logrus.Errorf("Failed to start sequence pool: %v", err)
				continue
			}
			sequencePools[*sequenceID] = true
		}
		
		// Queue message to appropriate pool
		var broadcastType, broadcastID string
		if msg.CampaignID != nil {
			broadcastType = "campaign"
			broadcastID = fmt.Sprintf("%d", *msg.CampaignID)
		} else if msg.SequenceID != nil {
			broadcastType = "sequence"
			broadcastID = *msg.SequenceID
		}
		
		if broadcastType != "" {
			err = p.manager.QueueMessageToBroadcast(broadcastType, broadcastID, &msg)
			if err != nil {
				logrus.Errorf("Failed to queue message: %v", err)
				// Update to failed
				db.Exec(` + "`" + `UPDATE broadcast_messages SET status = 'failed', error_message = ? WHERE id = ?` + "`" + `, 
					err.Error(), msg.ID)
			} else {
				messageCount++
			}
		}
	}
	
	if messageCount > 0 {
		// logrus.Infof("Queued %d messages to broadcast pools", messageCount)
	}
}`

	newProcessMessages := `func (p *UltraOptimizedBroadcastProcessor) processMessages() {
	broadcastRepo := repository.GetBroadcastRepository()
	
	// Get list of devices with pending messages
	devices, err := broadcastRepo.GetDevicesWithPendingMessages()
	if err != nil {
		logrus.Errorf("Failed to get devices with pending messages: %v", err)
		return
	}
	
	if len(devices) == 0 {
		return
	}
	
	logrus.Infof("Found %d devices with pending messages", len(devices))
	
	campaignPools := make(map[int]bool)
	sequencePools := make(map[string]bool)
	totalMessageCount := 0
	
	// Process each device
	for _, deviceID := range devices {
		// Use GetPendingMessagesAndLock for atomic message claiming
		messages, err := broadcastRepo.GetPendingMessagesAndLock(deviceID, 100)
		if err != nil {
			logrus.Errorf("Failed to get pending messages for device %s: %v", deviceID, err)
			continue
		}
		
		if len(messages) == 0 {
			continue
		}
		
		logrus.Infof("Processing %d messages for device %s", len(messages), deviceID)
		
		// Process each message
		for _, msg := range messages {
			// Create pool if needed
			if msg.CampaignID != nil && !campaignPools[*msg.CampaignID] {
				_, err := p.manager.StartBroadcastPool("campaign", fmt.Sprintf("%d", *msg.CampaignID))
				if err != nil {
					logrus.Errorf("Failed to start campaign pool: %v", err)
					continue
				}
				campaignPools[*msg.CampaignID] = true
				
				// Update campaign status to processing
				db := database.GetDB()
				db.Exec(` + "`" + `UPDATE campaigns SET status = 'processing', 
						 updated_at = NOW() 
						 WHERE id = ?` + "`" + `, *msg.CampaignID)
			}
			
			if msg.SequenceID != nil && !sequencePools[*msg.SequenceID] {
				_, err := p.manager.StartBroadcastPool("sequence", *msg.SequenceID)
				if err != nil {
					logrus.Errorf("Failed to start sequence pool: %v", err)
					continue
				}
				sequencePools[*msg.SequenceID] = true
			}
			
			// Queue message to appropriate pool
			var broadcastType, broadcastID string
			if msg.CampaignID != nil {
				broadcastType = "campaign"
				broadcastID = fmt.Sprintf("%d", *msg.CampaignID)
			} else if msg.SequenceID != nil {
				broadcastType = "sequence"
				broadcastID = *msg.SequenceID
			}
			
			if broadcastType != "" {
				err = p.manager.QueueMessageToBroadcast(broadcastType, broadcastID, &msg)
				if err != nil {
					logrus.Errorf("Failed to queue message: %v", err)
					// Update to failed
					db := database.GetDB()
					db.Exec(` + "`" + `UPDATE broadcast_messages SET status = 'failed', error_message = ? WHERE id = ?` + "`" + `, 
						err.Error(), msg.ID)
				} else {
					totalMessageCount++
				}
			}
		}
	}
	
	if totalMessageCount > 0 {
		logrus.Infof("Queued %d messages to broadcast pools", totalMessageCount)
	}
}`

	// Replace the function
	fileContent = strings.Replace(fileContent, oldProcessMessages, newProcessMessages, 1)

	// Add import for repository if not already present
	if !strings.Contains(fileContent, `"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"`) {
		// Find the import block and add repository
		importIndex := strings.Index(fileContent, `import (`)
		if importIndex != -1 {
			endImportIndex := strings.Index(fileContent[importIndex:], ")")
			if endImportIndex != -1 {
				endImportIndex += importIndex
				imports := fileContent[importIndex:endImportIndex]
				newImports := imports + "\t\"github.com/aldinokemal/go-whatsapp-web-multidevice/repository\"\n"
				fileContent = fileContent[:importIndex] + newImports + fileContent[endImportIndex:]
			}
		}
	}

	// Write the fixed content back
	err = ioutil.WriteFile(filePath, []byte(fileContent), 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	fmt.Println("Successfully updated UltraOptimizedBroadcastProcessor to use GetPendingMessagesAndLock!")
}
