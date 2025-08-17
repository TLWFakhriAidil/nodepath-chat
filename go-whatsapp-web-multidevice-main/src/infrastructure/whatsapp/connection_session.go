package whatsapp

import (
	"sync"
)

// ConnectionSession stores information about a device connection attempt
type ConnectionSession struct {
	UserID   string
	DeviceID string
}

var (
	// Change to store by deviceID instead of userID to support multiple devices per user
	connectionSessions = make(map[string]*ConnectionSession) // deviceID -> session
	sessionMutex       sync.RWMutex
)

// StoreConnectionSession stores a connection session by device ID
func StoreConnectionSession(deviceID string, session *ConnectionSession) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	connectionSessions[deviceID] = session
}

// GetConnectionSession retrieves a connection session by device ID
func GetConnectionSession(deviceID string) *ConnectionSession {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()
	return connectionSessions[deviceID]
}

// GetConnectionSessionByPhone finds a session that might match this phone
// This is used as a fallback when device ID is not known
func GetConnectionSessionByPhone(phone string) *ConnectionSession {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()
	
	// Return the most recent session (temporary solution)
	// In a proper implementation, we'd need to match based on some criteria
	for _, session := range connectionSessions {
		return session // Return first found
	}
	return nil
}

// GetAllConnectionSessions returns all active sessions
func GetAllConnectionSessions() map[string]*ConnectionSession {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()
	
	// Create a copy
	sessions := make(map[string]*ConnectionSession)
	for k, v := range connectionSessions {
		sessions[k] = v
	}
	return sessions
}

// ClearConnectionSession removes a connection session by device ID
func ClearConnectionSession(deviceID string) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	delete(connectionSessions, deviceID)
}

// ClearConnectionSessionByUserID removes all sessions for a user
func ClearConnectionSessionByUserID(userID string) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	
	// Find and remove all sessions for this user
	for deviceID, session := range connectionSessions {
		if session != nil && session.UserID == userID {
			delete(connectionSessions, deviceID)
		}
	}
}