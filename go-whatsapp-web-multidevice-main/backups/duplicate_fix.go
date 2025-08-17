// Add this check before sending any message
func (dw *DeviceWorker) sendMessage(msg domainBroadcast.BroadcastMessage) error {
    // CRITICAL: Check if message was already sent
    db := database.GetDB()
    var status string
    err := db.QueryRow("SELECT status FROM broadcast_messages WHERE id = ?", msg.ID).Scan(&status)
    if err == nil && status == "sent" {
        logrus.Warnf("Message %s already sent, skipping duplicate send", msg.ID)
        return nil // Don't send again
    }
    
    // Continue with normal sending...
}