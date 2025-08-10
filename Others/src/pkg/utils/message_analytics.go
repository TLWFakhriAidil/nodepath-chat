package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"
)

// MessageRecord represents a WhatsApp message for analytics
type MessageRecord struct {
	ID         string
	UserEmail  string // User email who owns this message
	JID        string
	Content    string
	Timestamp  time.Time
	FromMe     bool
	Status     string // sent, delivered, read, replied
	DeviceID   string
	DeviceName string
}

var (
	messageRecords     []MessageRecord
	messageRecordsMux  sync.RWMutex
	messageRecordsFile = "storages/message_records.csv"
)

// InitMessageRecords initializes the message records storage
func InitMessageRecords() error {
	// Create storages directory if it doesn't exist
	if err := os.MkdirAll("storages", 0755); err != nil {
		return err
	}
	
	// Load existing records from CSV if file exists
	if _, err := os.Stat(messageRecordsFile); err == nil {
		return loadMessageRecordsFromCSV()
	}
	
	return nil
}

// RecordMessageForUser records a message for a specific user
func RecordMessageForUser(id, userEmail, jid, content string, fromMe bool, status, deviceID, deviceName string) {
	messageRecordsMux.Lock()
	defer messageRecordsMux.Unlock()
	
	record := MessageRecord{
		ID:         id,
		UserEmail:  userEmail,
		JID:        jid,
		Content:    content,
		Timestamp:  time.Now(),
		FromMe:     fromMe,
		Status:     status,
		DeviceID:   deviceID,
		DeviceName: deviceName,
	}
	
	messageRecords = append(messageRecords, record)
	
	// Append to CSV file
	go appendMessageRecordToCSV(record)
}

// UpdateMessageStatus updates the status of an existing message
func UpdateMessageStatus(messageID, newStatus string) {
	messageRecordsMux.Lock()
	defer messageRecordsMux.Unlock()
	
	for i := range messageRecords {
		if messageRecords[i].ID == messageID {
			messageRecords[i].Status = newStatus
			// Update in CSV
			go saveAllMessageRecordsToCSV()
			break
		}
	}
}

// GetUserAnalytics returns analytics for a specific user
func GetUserAnalytics(userEmail string, startDate, endDate time.Time, deviceFilter string) map[string]interface{} {
	messageRecordsMux.RLock()
	defer messageRecordsMux.RUnlock()
	
	// Initialize counters
	leadsSent := 0
	leadsDelivered := 0
	leadsRead := 0
	leadsReplied := 0
	uniqueChats := make(map[string]bool)
	deviceStats := make(map[string]map[string]int)
	activeDevices := make(map[string]bool)
	
	// Daily stats
	dailyStats := make(map[string]map[string]int)
	
	// Initialize daily stats map
	for d := startDate; d.Before(endDate.Add(24 * time.Hour)); d = d.Add(24 * time.Hour) {
		dateStr := d.Format("2006-01-02")
		dailyStats[dateStr] = map[string]int{
			"sent":      0,
			"delivered": 0,
			"read":      0,
			"replied":   0,
		}
	}
	
	for _, record := range messageRecords {
		// Filter by user email
		if record.UserEmail != userEmail {
			continue
		}
		
		// Filter by date range
		if record.Timestamp.Before(startDate) || record.Timestamp.After(endDate.Add(24*time.Hour)) {
			continue
		}
		
		// Filter by device if specified
		if deviceFilter != "all" && deviceFilter != "" && record.DeviceID != deviceFilter {
			continue
		}
		
		// Count unique chats
		uniqueChats[record.JID] = true
		
		// Track active devices
		if record.DeviceName != "" {
			activeDevices[record.DeviceID] = true
			
			// Initialize device stats if not exists
			if _, exists := deviceStats[record.DeviceID]; !exists {
				deviceStats[record.DeviceID] = map[string]int{
					"sent":      0,
					"delivered": 0,
					"read":      0,
					"replied":   0,
					"name":      0, // Will store device name separately
				}
			}
		}
		
		dateStr := record.Timestamp.Format("2006-01-02")
		
		// Count based on message direction and status
		if record.FromMe {
			// Outgoing message (lead sent)
			leadsSent++
			dailyStats[dateStr]["sent"]++
			if record.DeviceID != "" {
				deviceStats[record.DeviceID]["sent"]++
			}
			
			// Check status
			switch record.Status {
			case "delivered":
				leadsDelivered++
				dailyStats[dateStr]["delivered"]++
				if record.DeviceID != "" {
					deviceStats[record.DeviceID]["delivered"]++
				}
			case "read":
				leadsDelivered++
				leadsRead++
				dailyStats[dateStr]["delivered"]++
				dailyStats[dateStr]["read"]++
				if record.DeviceID != "" {
					deviceStats[record.DeviceID]["delivered"]++
					deviceStats[record.DeviceID]["read"]++
				}
			}
		} else {
			// Incoming message (potential reply)
			leadsReplied++
			dailyStats[dateStr]["replied"]++
			if record.DeviceID != "" {
				deviceStats[record.DeviceID]["replied"]++
			}
		}
	}
	
	// Calculate percentages
	leadsNotReceived := leadsSent - leadsDelivered
	leadsNotRead := leadsDelivered - leadsRead
	
	// Convert daily stats to array format
	var dailyArray []map[string]interface{}
	for d := startDate; d.Before(endDate.Add(24 * time.Hour)); d = d.Add(24 * time.Hour) {
		dateStr := d.Format("2006-01-02")
		if stats, exists := dailyStats[dateStr]; exists {
			dailyArray = append(dailyArray, map[string]interface{}{
				"date":      d.Format("Jan 2"),
				"sent":      stats["sent"],
				"delivered": stats["delivered"],
				"read":      stats["read"],
				"replied":   stats["replied"],
			})
		}
	}
	
	// Get all user devices
	userDevices := GetUserDevices(userEmail)
	activeDeviceCount := 0
	inactiveDeviceCount := 0
	
	for _, device := range userDevices {
		if _, isActive := activeDevices[device["id"].(string)]; isActive {
			activeDeviceCount++
		} else {
			inactiveDeviceCount++
		}
	}
	
	return map[string]interface{}{
		"metrics": map[string]interface{}{
			"activeDevices":     activeDeviceCount,
			"inactiveDevices":   inactiveDeviceCount,
			"leadsSent":         leadsSent,
			"leadsReceived":     leadsDelivered,
			"leadsNotReceived":  leadsNotReceived,
			"leadsRead":         leadsRead,
			"leadsNotRead":      leadsNotRead,
			"leadsReplied":      leadsReplied,
		},
		"daily":       dailyArray,
		"deviceStats": deviceStats,
	}
}

// GetUserDevices returns all devices for a user
func GetUserDevices(userEmail string) []map[string]interface{} {
	// This would be implemented to get devices from database
	// For now, return mock data
	return []map[string]interface{}{
		{
			"id":     "device1",
			"name":   "iPhone 12",
			"status": "online",
		},
	}
}

// Helper functions for CSV operations
func loadMessageRecordsFromCSV() error {
	file, err := os.Open(messageRecordsFile)
	if err != nil {
		return err
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	
	messageRecordsMux.Lock()
	defer messageRecordsMux.Unlock()
	
	messageRecords = []MessageRecord{}
	for _, record := range records {
		if len(record) >= 8 {
			timestamp, _ := time.Parse(time.RFC3339, record[3])
			fromMe := record[4] == "true"
			
			messageRecords = append(messageRecords, MessageRecord{
				ID:         record[0],
				UserEmail:  record[1],
				JID:        record[2],
				Content:    record[3],
				Timestamp:  timestamp,
				FromMe:     fromMe,
				Status:     record[5],
				DeviceID:   record[6],
				DeviceName: record[7],
			})
		}
	}
	
	return nil
}

func appendMessageRecordToCSV(record MessageRecord) error {
	file, err := os.OpenFile(messageRecordsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	return writer.Write([]string{
		record.ID,
		record.UserEmail,
		record.JID,
		record.Content,
		record.Timestamp.Format(time.RFC3339),
		fmt.Sprintf("%v", record.FromMe),
		record.Status,
		record.DeviceID,
		record.DeviceName,
	})
}

func saveAllMessageRecordsToCSV() error {
	file, err := os.Create(messageRecordsFile)
	if err != nil {
		return err
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	for _, record := range messageRecords {
		err := writer.Write([]string{
			record.ID,
			record.UserEmail,
			record.JID,
			record.Content,
			record.Timestamp.Format(time.RFC3339),
			fmt.Sprintf("%v", record.FromMe),
			record.Status,
			record.DeviceID,
			record.DeviceName,
		})
		if err != nil {
			return err
		}
	}
	
	return nil
}