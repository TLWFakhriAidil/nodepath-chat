package rest

import (
    "go-whatsapp-web-multidevice/src/infrastructure/database"
    "go-whatsapp-web-multidevice/src/infrastructure/whatsapp"
    "log"
    "net/http"
    
    "github.com/gin-gonic/gin"
)

// Enhanced login handler that clears old session before creating new one
func EnhancedLoginHandler(c *gin.Context) {
    userID := c.Param("user")
    deviceID := c.Query("deviceId")
    
    if deviceID == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "code":    "BAD_REQUEST",
            "message": "Device ID is required",
        })
        return
    }
    
    log.Printf("Login request - UserID: %s, DeviceID: %s", userID, deviceID)
    
    // Clear any existing WhatsApp session first
    err := whatsapp.ClearWhatsAppSession(deviceID)
    if err != nil {
        log.Printf("Warning: Failed to clear old session: %v", err)
        // Continue with login anyway
    }
    
    // Update device status to connecting
    db := database.DBConn
    _, err = db.Exec(`
        UPDATE devices 
        SET status = 'connecting', 
            phone = NULL, 
            jid = NULL,
            updated_at = CURRENT_TIMESTAMP 
        WHERE id = $1
    `, deviceID)
    
    if err != nil {
        log.Printf("Error updating device status: %v", err)
    }
    
    // Continue with normal login process
    // ... (rest of the existing login code)
}

// Fix for reconnection after logout
func HandleReconnect(c *gin.Context) {
    deviceID := c.Query("deviceId")
    
    if deviceID == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "code":    "BAD_REQUEST",
            "message": "Device ID is required",
        })
        return
    }
    
    log.Printf("Reconnect request for device: %s", deviceID)
    
    // First ensure device is properly logged out
    err := whatsapp.HandleDeviceLogout(deviceID)
    if err != nil {
        log.Printf("Error during logout: %v", err)
    }
    
    // Clear all session data
    err = whatsapp.ClearWhatsAppSession(deviceID)
    if err != nil {
        log.Printf("Error clearing session: %v", err)
    }
    
    // Now proceed with new login
    c.JSON(http.StatusOK, gin.H{
        "code":    "SUCCESS",
        "message": "Device ready for new connection",
    })
}
